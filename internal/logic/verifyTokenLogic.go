package logic

import (
	"context"

	"auth-service/internal/svc"
	"auth-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyTokenLogic {
	return &VerifyTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyTokenLogic) VerifyToken(req *types.VerifyTokenReq) (resp *types.VerifyTokenResp, err error) {
	// 验证访问令牌
	claims, err := l.svcCtx.JWT.VerifyAccessToken(req.Token)
	if err != nil {
		return &types.VerifyTokenResp{
			Valid: false,
		}, nil
	}
	
	return &types.VerifyTokenResp{
		Valid:     true,
		UserID:    int64(claims.UserID),
		Username:  claims.Username,
		ExpiresAt: claims.ExpiresAt,
	}, nil
}
