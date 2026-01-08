package logic

import (
	"context"

	"auth-service/internal/svc"
	"auth-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshLogic) Refresh(req *types.RefreshReq) (resp *types.RefreshResp, err error) {
	l.Info("Refresh request received")

	// 使用 JWT 服务刷新令牌
	tokenPair, err := l.svcCtx.JWT.Refresh(req.RefreshToken)
	if err != nil {
		l.Errorf("Failed to refresh token: %v", err)
		return nil, types.ErrInvalidRefreshToken
	}

	// 返回新的令牌对
	resp = &types.RefreshResp{
		AccessToken:      tokenPair.AccessToken,
		AccessExpiresAt:  tokenPair.AccessExpiresAt,
		RefreshToken:     tokenPair.RefreshToken,
		RefreshExpiresAt: tokenPair.RefreshExpiresAt,
		TokenType:        "Bearer",
	}

	l.Infof("Token refreshed successfully")
	return resp, nil
}
