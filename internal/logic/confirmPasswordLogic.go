package logic

import (
	"context"

	"auth-service/internal/svc"
	"auth-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmPasswordLogic {
	return &ConfirmPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmPasswordLogic) ConfirmPassword(req *types.ConfirmPasswordReq) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
