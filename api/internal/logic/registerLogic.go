package logic

import (
	"context"
	"time"

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
	l.Info("Register request received", "username", req.Username, "email", req.Email)

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

	// 检查用户名和邮箱是否已存在
	var (
		userByUsername *model.User
		userByEmail    *model.User
		errUsername    error
		errEmail       error
	)

	// 使用 mr.Finish 优雅并发查询
	mr.Finish(
		func() error {
			errUsername = l.svcCtx.DB.QueryRowCtx(l.ctx, &userByUsername, "SELECT * FROM user WHERE username = ?", req.Username)
			if errUsername != nil && errUsername != model.ErrNotFound {
				l.Infof("Find user by username error: %v", errUsername)
				return errUsername
			}
			return nil
		},
		func() error {
			errEmail = l.svcCtx.DB.QueryRowCtx(l.ctx, &userByEmail, "SELECT * FROM user WHERE email = ?", req.Email)
			if errEmail != nil && errEmail != model.ErrNotFound {
				l.Infof("Find user by email error: %v", errEmail)
				return errEmail
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

	// 构建用户模型
	now := time.Now().Unix()
	newUser := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Nickname: req.Nickname,
		PasswordHash: l.svcCtx.PasswordEncoder.Hash(req.Password),
		AccountStatus:       model.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 5. 插入数据库
	if err := l.svcCtx.DB.RawDB().Insert(newUser); err != nil {
		l.Errorf("Failed to insert user: %v", err)
		return nil, types.ErrDatabaseError
	}

	// 6. 打印成功日志
	l.Infof("User registered successfully: userId=%s, username=%s", newUser.UserID, req.Username)

	// 7. 返回响应
	resp := &types.RegisterResp{
		UserID:    newUser.UserID,
		Username:  newUser.Username,
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt,
	}

	return resp, nil
}
