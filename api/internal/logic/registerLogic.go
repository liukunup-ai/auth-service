package logic

import (
	"context"
	"time"

	"auth-service/api/internal/svc"
	"auth-service/api/internal/types"
	"auth-service/model"

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
	// 1. 日志记录开始
	l.Info("Register request received", "username", req.Username, "email", req.Email)

	// 2. 校验验证码（如果开启了验证码）
	if l.svcCtx.Config.Captcha.Enable && (req.CaptchaID != "" || req.CaptchaAnswer != "") {
		match := l.svcCtx.Captcha.Verify(req.CaptchaID, req.CaptchaAnswer, true)
		if !match {
			l.Info("Captcha verify failed", "captchaId", req.CaptchaID)
			return nil, types.ErrCaptchaInvalid // 自定义错误
		}
	}

	// 3. 并发检查：用户名和邮箱是否已存在
	var (
		userByUsername *model.User
		userByEmail    *model.User
		errUsername    error
		errEmail       error
	)

	// 使用 mr.Finish 优雅并发查询
	mr.Finish(
		func() {
			userByUsername, errUsername = l.svcCtx.UserModel.FindByUsername(l.ctx, req.Username)
			if errUsername != nil && errUsername != model.ErrNotFound {
				l.Errorf("Find user by username error: %v", errUsername)
			}
		},
		func() {
			userByEmail, errEmail = l.svcCtx.UserModel.FindByEmail(l.ctx, req.Email)
			if errEmail != nil && errEmail != model.ErrNotFound {
				l.Errorf("Find user by email error: %v", errEmail)
			}
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

	// 4. 构建用户模型
	now := time.Now().Unix()
	newUser := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Nickname: req.Nickname,
		// 密码加密
		PasswordHash: l.svcCtx.PasswordEncoder.Hash(req.Password),
		Status:       model.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 5. 插入数据库
	if err := l.svcCtx.UserModel.Insert(l.ctx, newUser); err != nil {
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
