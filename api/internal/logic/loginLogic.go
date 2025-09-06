package logic

import (
	"context"

	"auth-service/api/internal/svc"
	"auth-service/api/internal/types"
	model "auth-service/model/mysql"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	l.Info("Login request received", ", username: ", req.Username, ", captchaId: ", req.CaptchaID)

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

	var (
		user           *model.User
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

	// 检查是否存在
	if userByUsername != nil { // found by username
		user = userByUsername
	} else if userByEmail != nil { // found by email
		user = userByEmail
	} else if userByPhone != nil { // found by phone
		user = userByPhone
	} else {
		l.Info("User not found", "username", req.Username)
		return nil, types.ErrUserNotFound
	}

	// 校验密码
	if l.svcCtx.PasswordEncoder.Compare(user.PasswordHash, req.Password) {
		l.Info("Invalid password for user", "username", req.Username)
		return nil, types.ErrInvalidPassword
	}

	// 生成 JWT Pair
	tokenPair, err := l.svcCtx.JWT.Generate(user.Id, user.Username)
	if err != nil {
		l.Errorf("Failed to generate JWT tokens: %v", err)
		return nil, types.ErrGenerateToken
	}

	resp = &types.LoginResp{
		UserID:           user.PublicId,
		Username:         user.Username,
		Email:            user.Email,
		AccessToken:      tokenPair.AccessToken,
		AccessExpiresIn:  tokenPair.AccessExpire,
		RefreshToken:     tokenPair.RefreshToken,
		RefreshExpiresIn: tokenPair.RefreshExpire,
		TokenType:        "Bearer",
	}
	return resp, nil
}
