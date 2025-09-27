package logic

import (
	"context"

	"auth-service/internal/svc"
	"auth-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenReq) (resp *types.RefreshTokenResp, err error) {
	// 刷新
	tokenPair, err := l.svcCtx.JWT.Refresh(req.RefreshToken)
	if err != nil {
		return &types.RefreshTokenResp{}, err
	}

	resp = &types.RefreshTokenResp{
		AccessToken:      tokenPair.AccessToken,
		AccessExpiresAt:  tokenPair.AccessExpiresAt,
		RefreshToken:     tokenPair.RefreshToken,
		RefreshExpiresAt: tokenPair.RefreshExpiresAt,
		TokenType:        "Bearer",
	}
	return resp, nil
}
