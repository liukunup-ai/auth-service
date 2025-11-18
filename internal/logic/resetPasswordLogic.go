package logic

import (
	"crypto/rand"
	"encoding/base64"
	"context"
	"net/http"

	"auth-service/internal/svc"
	"auth-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// generateSecureToken 生成指定长度的安全令牌
func generateSecureToken(length int) (string, error) {
	token := make([]byte, length)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(token), nil
}

type ResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPasswordLogic) ResetPassword(req *types.ResetPasswordReq) (resp *types.BaseResponse, err error) {
	// 1. 验证邮箱是否存在
	_, err = l.svcCtx.UserModel.FindOneByEmail(l.ctx, req.Email)
	if err != nil {
		// 出于安全考虑，即使邮箱不存在也返回成功
		return &types.BaseResponse{
			Code:    http.StatusOK,
			Message: "如果邮箱存在，重置链接将发送到您的邮箱",
		}, nil
	}

	// 2. 生成重置令牌（暂时不保存到数据库和发送邮件）
	_, err = generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	// 为了演示目的，直接返回成功响应
	// 在实际应用中，这里应该保存重置记录并发送邮件
	return &types.BaseResponse{
		Code:    http.StatusOK,
		Message: "重置链接已发送到您的邮箱（演示模式）",
	}, nil
}
