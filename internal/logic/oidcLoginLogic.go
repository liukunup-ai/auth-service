// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"fmt"
	"time"

	"auth-service/internal/svc"
	"auth-service/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type OIDCLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOIDCLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OIDCLoginLogic {
	return &OIDCLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OIDCLoginLogic) OIDCLogin(req *types.OIDCLoginReq) (resp *types.BaseResponse, err error) {
	if !l.svcCtx.OIDC.IsEnabled() {
		return &types.BaseResponse{
			Code:    1001,
			Message: "OIDC login is disabled",
		}, nil
	}

	state := uuid.New().String()
	nonce := uuid.New().String()

	// Cache state to verify callback later
	key := fmt.Sprintf("auth:oidc:state:%s", state)
	redirectURL := req.RedirectURL
	if redirectURL == "" {
		redirectURL = "/"
	}

	err = l.svcCtx.Redis.Set(l.ctx, key, redirectURL, 5*time.Minute).Err()
	if err != nil {
		l.Logger.Errorf("failed to cache OIDC state: %v", err)
		return &types.BaseResponse{
			Code:    500,
			Message: "internal server error",
		}, nil
	}

	authURL := l.svcCtx.OIDC.GetAuthorizationURL(state, nonce)

	return &types.BaseResponse{
		Code:    0,
		Message: "success",
		Data: types.OIDCLoginResp{
			AuthorizationURL: authURL,
			State:            state,
		},
	}, nil
}
