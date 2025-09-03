package logic

import (
	"context"

	"auth-service/api/internal/svc"
	"auth-service/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthCheckLogic {
	return &HealthCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthCheckLogic) HealthCheck() (resp *types.BaseResponse, err error) {
	resp = &types.BaseResponse{
		Code:    200,
		Message: "Service is healthy",
		Data:    map[string]interface{}{"status": "ok"},
	}

	logx.Infof("HealthCheck passed at %v", l.ctx)
	return resp, nil
}
