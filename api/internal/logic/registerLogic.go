package logic

import (
	"context"
	"database/sql"

	"auth-service/api/internal/svc"
	"auth-service/api/internal/types"
	model "auth-service/model/mysql"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	l.Info("Register request received", "username", req.Username, "email", req.Email, "phone", req.Phone)

	// 校验验证码（如果开启了验证码）
	if l.svcCtx.Config.Captcha.Enable {
		if req.CaptchaID != "" || req.CaptchaAnswer != "" {
			match := l.svcCtx.Captcha.Verify(req.CaptchaID, req.CaptchaAnswer, true)
			if !match {
				l.Info("Captcha verify failed", "captchaId", req.CaptchaID)
				return nil, types.ErrCaptchaInvalid
			}
		} else {
			l.Info("Captcha is required but not provided")
			return nil, types.ErrCaptchaRequired
		}
	}

	// 检查用户名/邮箱/手机号是否已存在
	var (
		userByUsername *model.User
		userByEmail    *model.User
		userByPhone    *model.User
		errUsername    error
		errEmail       error
		errPhone       error
	)

	// 使用 mr.Finish 优雅并发查询
	mr.Finish(
		func() error { // 查询用户名
			errUsername = l.svcCtx.DB.QueryRowCtx(l.ctx, &userByUsername, "SELECT * FROM user WHERE username = ?", req.Username)
			if errUsername != nil && errUsername != model.ErrNotFound {
				l.Infof("Find user by username error: %v", errUsername)
				return errUsername
			}
			return nil
		},
		func() error { // 查询邮箱
			errEmail = l.svcCtx.DB.QueryRowCtx(l.ctx, &userByEmail, "SELECT * FROM user WHERE email = ?", req.Username)
			if errEmail != nil && errEmail != model.ErrNotFound {
				l.Infof("Find user by email error: %v", errEmail)
				return errEmail
			}
			return nil
		},
		func() error { // 查询手机号
			errPhone = l.svcCtx.DB.QueryRowCtx(l.ctx, &userByPhone, "SELECT * FROM user WHERE phone = ?", req.Username)
			if errPhone != nil && errPhone != model.ErrNotFound {
				l.Infof("Find user by phone error: %v", errPhone)
				return errPhone
			}
			return nil
		},
	)

	// 检查是否已存在
	if userByUsername != nil {
		l.Info("Username already exists", "username", req.Username)
		return nil, types.ErrUsernameTaken
	}
	if userByEmail != nil {
		l.Info("Email already exists", "email", req.Email)
		return nil, types.ErrEmailTaken
	}
	if userByPhone != nil {
		l.Info("Phone already exists", "phone", req.Phone)
		return nil, types.ErrPhoneTaken
	}

	// 构建用户模型
	newUser := &model.User{
		Username:      req.Username,
		Email:         req.Email,
		Phone:         sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		Nickname:      sql.NullString{String: req.Nickname, Valid: req.Nickname != ""},
		PasswordHash:  l.svcCtx.PasswordEncoder.Hash(req.Password),
		AccountStatus: model.UserStatusActive,
	}

	// 插入数据库
	result, err := l.svcCtx.UserModel.Insert(l.ctx, newUser)
	if err != nil {
		l.Errorf("Failed to insert user into database: %v", err)
		return nil, types.ErrDatabaseError
	}
	if n, err := result.RowsAffected(); err != nil || n == 0 {
		l.Errorf("No rows affected when inserting user: %v, rows affected: %d", err, n)
		return nil, types.ErrDatabaseError
	}

	// 打印成功日志
	l.Infof("User registered successfully: %s (ID: %d)", newUser.Username, newUser.PublicId)

	// 返回响应
	resp = &types.RegisterResp{
		UserID:    newUser.PublicId,
		Username:  newUser.Username,
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt.Unix(),
	}

	return resp, nil
}
