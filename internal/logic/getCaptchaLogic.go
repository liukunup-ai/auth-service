package logic

import (
	"context"

	"auth-service/internal/svc"
	"auth-service/internal/types"

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

	resp = &types.CaptchaResp{
		CaptchaID:    id,
		CaptchaImage: b64s,
		ExpiresIn:    l.svcCtx.Config.Captcha.ExpiresIn,
	}
	return resp, nil
}
