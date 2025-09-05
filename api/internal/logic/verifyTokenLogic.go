package logic

import (
	"context"

	"auth-service/api/internal/svc"
	"auth-service/api/internal/types"

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
	// 验证令牌
	claims, err := l.svcCtx.JWT.VerifyAccessToken(req.Token)
	if err != nil {
		return &types.VerifyTokenResp{
			IsValid: false,
			Message: err.Error(),
		}, err
	}

	// 查询用户信息
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, claims.UserID)
	if err != nil {
		return &types.VerifyTokenResp{
			IsValid: false,
			Message: "用户不存在",
		}, nil
	}

	resp = &types.VerifyTokenResp{
		IsValid:   true,
		UserID:    user.PublicId,
		Username:  claims.Username,
		ExpiresAt: claims.ExpiresAt,
	}
	return resp, nil
}
