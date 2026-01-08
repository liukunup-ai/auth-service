package logic

import (
	"context"

	"auth-service/internal/svc"
	"auth-service/internal/types"

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
	l.Info("ChangePassword request received")

	// 校验新旧密码是否一致
	if req.NewPassword == req.OldPassword {
		return &types.BaseResponse{
			Code:    400,
			Message: "新密码和旧密码不能相同",
		}, nil
	}

	// 从上下文中获取用户ID（由中间件设置）
	userID, ok := l.ctx.Value("userID").(int64)
	if !ok || userID == 0 {
		l.Error("Failed to get userID from context")
		return &types.BaseResponse{
			Code:    401,
			Message: "未授权",
		}, nil
	}

	// 查询用户信息
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(userID))
	if err != nil {
		l.Errorf("Failed to find user by id %d: %v", userID, err)
		return &types.BaseResponse{
			Code:    404,
			Message: "用户不存在",
		}, nil
	}

	// 校验旧密码
	if !l.svcCtx.PasswordEncoder.Compare(user.PasswordHash, req.OldPassword) {
		l.Info("Old password verification failed")
		return &types.BaseResponse{
			Code:    400,
			Message: "旧密码错误",
		}, nil
	}

	// 更新密码
	user.PasswordHash = l.svcCtx.PasswordEncoder.Hash(req.NewPassword)
	err = l.svcCtx.UserModel.Update(l.ctx, user)
	if err != nil {
		l.Errorf("Failed to update password: %v", err)
		return &types.BaseResponse{
			Code:    500,
			Message: "密码更新失败",
		}, nil
	}

	l.Infof("Password changed successfully for user %s", user.PublicId)
	return &types.BaseResponse{
		Code:    200,
		Message: "密码修改成功",
	}, nil
}
