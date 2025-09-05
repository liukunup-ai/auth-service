package logic

import (
	"context"

	"auth-service/api/internal/svc"
	"auth-service/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordReq) (resp *types.BaseResponse, err error) {
	// 校验新旧密码是否一致
	if req.NewPassword == req.OldPassword {
		return &types.BaseResponse{
			Code:    400,
			Message: "新密码和旧密码不能相同",
		}, nil
	}

	// 获取当前用户ID
	publicID, ok := l.ctx.Value("userId").(string)
	if !ok || publicID == "" {
		return &types.BaseResponse{
			Code:    401,
			Message: "未授权",
		}, nil
	}

	// 查询用户信息
	user, err := l.svcCtx.UserModel.FindOneByPublicId(l.ctx, publicID)
	if err != nil {
		return &types.BaseResponse{
			Code:    404,
			Message: "用户不存在",
		}, nil
	}

	// 校验旧密码
	if !l.svcCtx.PasswordEncoder.Compare(user.PasswordHash, req.OldPassword) {
		return &types.BaseResponse{
			Code:    400,
			Message: "旧密码错误",
		}, nil
	}

	// 更新密码
	user.PasswordHash = l.svcCtx.PasswordEncoder.Hash(req.NewPassword)
	err = l.svcCtx.UserModel.Update(l.ctx, user)
	if err != nil {
		return &types.BaseResponse{
			Code:    500,
			Message: "密码更新失败",
		}, nil
	}

	return &types.BaseResponse{
		Code:    200,
		Message: "密码修改成功",
	}, nil
}
