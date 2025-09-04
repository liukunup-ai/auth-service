package logic

import (
	"context"

	"auth-service/rpc/auth"
	"auth-service/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthCheckLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHealthCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthCheckLogic {
	return &HealthCheckLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 健康检查
func (l *HealthCheckLogic) HealthCheck(in *auth.Empty) (*auth.BaseResponse, error) {
	// todo: add your logic here and delete this line

	return &auth.BaseResponse{}, nil
}
