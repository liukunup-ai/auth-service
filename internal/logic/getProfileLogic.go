package logic

import (
	"context"

	"auth-service/internal/svc"
	"auth-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProfileLogic) GetProfile() (resp *types.UserProfileResp, err error) {
	l.Info("GetProfile request received")

	// 从上下文中获取用户ID（由中间件设置）
	userID, ok := l.ctx.Value("userID").(int64)
	if !ok || userID == 0 {
		l.Error("Failed to get userID from context")
		return nil, types.ErrUnauthorized
	}

	// 根据内部ID查询用户信息
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(userID))
	if err != nil {
		l.Errorf("Failed to find user by id %d: %v", userID, err)
		return nil, types.ErrUserNotFound
	}

	// 构造响应
	resp = &types.UserProfileResp{
		UserID:    user.PublicId,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone.String,
		Nickname:  user.Nickname.String,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}

	l.Infof("User profile retrieved successfully for user %s", user.PublicId)
	return resp, nil
}
