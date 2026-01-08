// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"auth-service/internal/svc"
	"auth-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ForgotPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewForgotPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForgotPasswordLogic {
	return &ForgotPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ForgotPasswordLogic) ForgotPassword(req *types.ForgotPasswordReq) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
