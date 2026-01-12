// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"database/sql"
	"fmt"

	"auth-service/internal/svc"
	"auth-service/internal/types"
	"auth-service/model/mysql"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type OIDCCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOIDCCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OIDCCallbackLogic {
	return &OIDCCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OIDCCallbackLogic) OIDCCallback(req *types.OIDCCallbackReq) (resp *types.BaseResponse, err error) {
	if !l.svcCtx.OIDC.IsEnabled() {
		return &types.BaseResponse{
			Code:    1001,
			Message: "OIDC login is disabled",
		}, nil
	}

	// 1. Verify State
	cachedKey := fmt.Sprintf("auth:oidc:state:%s", req.State)
	_, err = l.svcCtx.Redis.Get(l.ctx, cachedKey).Result()
	if err == redis.Nil {
		return &types.BaseResponse{
			Code:    1002,
			Message: "invalid or expired state",
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check state: %w", err)
	}
	defer l.svcCtx.Redis.Del(l.ctx, cachedKey)

	if req.Error != "" {
		return &types.BaseResponse{
			Code:    1003,
			Message: fmt.Sprintf("OIDC login failed: %s - %s", req.Error, req.ErrorDescription),
		}, nil
	}

	// 2. Exchange Code for Token
	tokenResp, err := l.svcCtx.OIDC.ExchangeCode(l.ctx, req.Code)
	if err != nil {
		l.Logger.Errorf("Failed to exchange code: %v", err)
		return &types.BaseResponse{
			Code:    1004,
			Message: "Failed to exchange code",
		}, nil
	}

	// 3. Get User Info
	userInfo, err := l.svcCtx.OIDC.GetUserInfo(l.ctx, tokenResp.AccessToken)
	if err != nil {
		l.Logger.Errorf("Failed to get user info: %v", err)
		return &types.BaseResponse{
			Code:    1005,
			Message: "Failed to get user info",
		}, nil
	}

	// 4. Find or Create User
	var user *mysql.User
	var isNewUser bool

	// Try to find by Email
	if userInfo.Email != "" {
		user, err = l.svcCtx.UserModel.FindOneByEmail(l.ctx, userInfo.Email)
		if err != nil && err != mysql.ErrNotFound {
			return nil, fmt.Errorf("failed to find user by email: %w", err)
		}
	}

	if user == nil {
		// New User Logic
		isNewUser = true

		email := userInfo.Email
		if email == "" {
			// Generate placeholder email to satisfy unique constraint
			email = fmt.Sprintf("%s@no-email.placeholder", uuid.New().String())
		}

		// Determine Username
		username := userInfo.PreferredUsername
		if username == "" {
			username = email
		}
		if username == "" || userInfo.Email == "" { // if original email was empty, username might be set to placeholder email which is ugly but unique.
			username = "user_" + uuid.New().String()[:8]
		}

		// Ensure username uniqueness
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

		// Create Password
		newPass := uuid.New().String()
		hashedPass := l.svcCtx.PasswordEncoder.Hash(newPass)

		newUser := &mysql.User{
			PublicId:      uuid.New().String(),
			Username:      username,
			Email:         email,
			EmailVerified: 0,
			PasswordHash:  hashedPass,
			AccountStatus: mysql.UserStatusActive, // 1
		}

		if userInfo.EmailVerified {
			newUser.EmailVerified = 1
		}
		if userInfo.Name != "" {
			newUser.Nickname = sql.NullString{String: userInfo.Name, Valid: true}
		}

		res, err := l.svcCtx.UserModel.Insert(l.ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		uid, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}

		// Fetch the complete user object
		user, err = l.svcCtx.UserModel.FindOne(l.ctx, uint64(uid))
		if err != nil {
			return nil, err
		}
	} else {
		// Existing user logic (optional: update info)
		// For example, if email was not verified but now OIDC says it is
		/*
			if user.EmailVerified == 0 && userInfo.EmailVerified {
				user.EmailVerified = 1
				_ = l.svcCtx.UserModel.Update(l.ctx, user)
			}
		*/
	}

	// 5. Generate JWT
	tokenPair, err := l.svcCtx.JWT.Generate(user.Id, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &types.BaseResponse{
		Code:    0,
		Message: "success",
		Data: types.OIDCCallbackResp{
			UserID:           user.PublicId,
			Username:         user.Username,
			Email:            user.Email,
			AccessToken:      tokenPair.AccessToken,
			AccessExpiresAt:  tokenPair.AccessExpiresAt,
			RefreshToken:     tokenPair.RefreshToken,
			RefreshExpiresAt: tokenPair.RefreshExpiresAt,
			TokenType:        "Bearer",
			IsNewUser:        isNewUser,
			Provider:         "oidc",
		},
	}, nil
}
