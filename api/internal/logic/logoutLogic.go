package logic

import (
	"context"

	"auth-service/api/internal/svc"
	"auth-service/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(req *types.LogoutReq) (resp *types.BaseResponse, err error) {
	// 登出
	if err := l.svcCtx.JWT.Logout(req.AccessToken, req.RefreshToken); err != nil {
		return &types.BaseResponse{
			Code:    400,
			Message: "Logout failed",
		}, err
	}

	resp = &types.BaseResponse{
		Code:    0,
		Message: "Logout",
	}
	return resp, nil
}
