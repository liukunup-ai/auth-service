package logic

import (
	"context"

	"auth-service/internal/svc"
	"auth-service/internal/types"
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
				l.Info("Captcha verify failed", ", captchaId: ", req.CaptchaID)
				return nil, types.ErrCaptchaInvalid
			}
		} else {
			l.Info("Captcha is required but not provided")
			return nil, types.ErrCaptchaRequired
		}
	}

	var (
		user            model.User
		userByUsername  model.User
		userByEmail     model.User
		userByPhone     model.User
		foundByUsername bool
		foundByEmail    bool
		foundByPhone    bool
	)

	// 使用 mr.Finish 优雅并发查询
	mr.Finish(
		func() error { // 查询用户名
			err := l.svcCtx.DB.QueryRowCtx(l.ctx, &userByUsername, "SELECT * FROM user WHERE username = ?", req.Username)
			if err == nil {
				foundByUsername = true
			} else if err != model.ErrNotFound {
				l.Infof("Find user by username error: %v", err)
				return err
			}
			return nil
		},
		func() error { // 查询邮箱
			err := l.svcCtx.DB.QueryRowCtx(l.ctx, &userByEmail, "SELECT * FROM user WHERE email = ?", req.Username)
			if err == nil {
				foundByEmail = true
			} else if err != model.ErrNotFound {
				l.Infof("Find user by email error: %v", err)
				return err
			}
			return nil
		},
		func() error { // 查询手机号
			err := l.svcCtx.DB.QueryRowCtx(l.ctx, &userByPhone, "SELECT * FROM user WHERE phone = ?", req.Username)
			if err == nil {
				foundByPhone = true
			} else if err != model.ErrNotFound {
				l.Infof("Find user by phone error: %v", err)
				return err
			}
			return nil
		},
	)

	// 检查是否存在
	if foundByUsername { // found by username
		user = userByUsername
	} else if foundByEmail { // found by email
		user = userByEmail
	} else if foundByPhone { // found by phone
		user = userByPhone
	} else {
		l.Info("User not found", ", username: ", req.Username)
		return nil, types.ErrUserNotFound
	}

	// 校验密码
	if !l.svcCtx.PasswordEncoder.Compare(user.PasswordHash, req.Password) {
		l.Info("Invalid password for user", ", username: ", req.Username)
		return nil, types.ErrInvalidPassword
	}

	// 生成 JWT Pair
	tokenPair, err := l.svcCtx.JWT.Generate(user.Id, user.Username)
	if err != nil {
		l.Errorf("Failed to generate JWT: %v", err)
		return nil, types.ErrGenerateToken
	}

	// 更新最后登录时间
	_, err = l.svcCtx.DB.ExecCtx(l.ctx, "UPDATE user SET last_login_at = NOW() WHERE id = ?", user.Id)
	if err != nil {
		l.Errorf("Failed to update last login time for user %d: %v", user.Id, err)
	}

	// 返回响应
	resp = &types.LoginResp{
		UserID:           user.PublicId,
		Username:         user.Username,
		Email:            user.Email,
		AccessToken:      tokenPair.AccessToken,
		AccessExpiresAt:  tokenPair.AccessExpiresAt,
		RefreshToken:     tokenPair.RefreshToken,
		RefreshExpiresAt: tokenPair.RefreshExpiresAt,
		TokenType:        "Bearer",
	}
	return resp, nil
}
