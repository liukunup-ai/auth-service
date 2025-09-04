package logic

import (
	"context"

	"auth-service/api/internal/svc"
	"auth-service/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCaptchaLogic {
	return &GetCaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCaptchaLogic) GetCaptcha() (resp *types.CaptchaResp, err error) {
	// 生成验证码
	id, b64s, _, err := l.svcCtx.Captcha.Generate()
	if err != nil {
		return nil, err
	}

	// 返回响应
	resp = &types.CaptchaResp{
		CaptchaID:    id,
		CaptchaImage: b64s,
		ExpiresIn:    l.svcCtx.Config.Captcha.Expire,
	}
	return resp, nil
}
