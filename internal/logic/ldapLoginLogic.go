// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"auth-service/internal/svc"
	"auth-service/internal/types"
	"auth-service/model/mysql"

	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type LDAPLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLDAPLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LDAPLoginLogic {
	return &LDAPLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LDAPLoginLogic) LDAPLogin(req *types.LDAPLoginReq) (resp *types.BaseResponse, err error) {
	if !l.svcCtx.LDAP.IsEnabled() {
		return &types.BaseResponse{
			Code:    1001,
			Message: "LDAP login is disabled",
		}, nil
	}

	// 1. Verify Captcha
	if l.svcCtx.Config.Captcha.Enable {
		if req.CaptchaID != "" || req.CaptchaAnswer != "" {
			match := l.svcCtx.Captcha.Verify(req.CaptchaID, req.CaptchaAnswer, true)
			if !match {
				return nil, types.ErrCaptchaInvalid
			}
		} else {
			return nil, types.ErrCaptchaRequired
		}
	}

	// 2. Authenticate against LDAP
	userInfo, err := l.svcCtx.LDAP.Authenticate(l.ctx, req.Username, req.Password)
	if err != nil {
		l.Infof("LDAP authentication failed for %s: %v", req.Username, err)
		return &types.BaseResponse{
			Code:    1002,
			Message: "authentication failed",
		}, nil
	}

	// 3. Find or Create User
	var user *mysql.User
	var isNewUser bool

	// Try to find by Username first (LDAP username usually matches)
	if userInfo.Username != "" {
		user, err = l.svcCtx.UserModel.FindOneByUsername(l.ctx, userInfo.Username)
		if err != nil && err != mysql.ErrNotFound {
			return nil, fmt.Errorf("failed to find user by username: %w", err)
		}
	}

	// If not found by username, try email if available
	if user == nil && userInfo.Email != "" {
		user, err = l.svcCtx.UserModel.FindOneByEmail(l.ctx, userInfo.Email)
		if err != nil && err != mysql.ErrNotFound {
			return nil, fmt.Errorf("failed to find user by email: %w", err)
		}
	}

	if user == nil {
		// Create new user
		isNewUser = true

		username := userInfo.Username
		email := userInfo.Email

		if username == "" {
			username = email
		}
		if username == "" {
			username = "ldap_" + uuid.New().String()[:8]
		}

		if email == "" {
			email = fmt.Sprintf("%s@no-email.placeholder", uuid.New().String())
		}

		// Ensure username uniqueness (though unlikely collision if coming from LDAP)
		for {
			_, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, username)
			if err == mysql.ErrNotFound {
				break
			}
			if err != nil {
				return nil, err
			}
			username = fmt.Sprintf("%s_%s", username, uuid.New().String()[:4])
		}

		// Create random password
		newPass := uuid.New().String()
		hashedPass := l.svcCtx.PasswordEncoder.Hash(newPass)

		newUser := &mysql.User{
			PublicId:      uuid.New().String(),
			Username:      username,
			Email:         email,
			EmailVerified: 0,
			PasswordHash:  hashedPass,
			AccountStatus: mysql.UserStatusActive,
		}

		if userInfo.DisplayName != "" {
			newUser.Nickname = sql.NullString{String: userInfo.DisplayName, Valid: true}
		}

		res, err := l.svcCtx.UserModel.Insert(l.ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		uid, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}

		user, err = l.svcCtx.UserModel.FindOne(l.ctx, uint64(uid))
		if err != nil {
			return nil, err
		}
	}

	// 4. Generate Token
	tokenPair, err := l.svcCtx.JWT.Generate(user.Id, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 5. Build Response
	return &types.BaseResponse{
		Code:    0,
		Message: "success",
		Data: types.LDAPLoginResp{
			UserID:           user.PublicId,
			Username:         user.Username,
			Email:            user.Email,
			DisplayName:      userInfo.DisplayName,
			Groups:           userInfo.Groups,
			AccessToken:      tokenPair.AccessToken,
			AccessExpiresAt:  tokenPair.AccessExpiresAt,
			RefreshToken:     tokenPair.RefreshToken,
			RefreshExpiresAt: tokenPair.RefreshExpiresAt,
			TokenType:        "Bearer",
			IsNewUser:        isNewUser,
			Provider:         "ldap",
		},
	}, nil
}
