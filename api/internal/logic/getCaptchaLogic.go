package logic

import (
	"context"
	"time"

	"auth-service/api/internal/svc"
	"auth-service/api/internal/types"
	util "auth-service/util/captcha"

	"github.com/mojocn/base64Captcha"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	DefaultCaptchaExpire = 3 // minutes
)

type GetCaptchaLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	captcha *base64Captcha.Captcha
}

func NewGetCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCaptchaLogic {
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	store := util.NewRedisStore(ctx, svcCtx.Redis, "captcha:", DefaultCaptchaExpire*time.Minute)
	return &GetCaptchaLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		captcha: base64Captcha.NewCaptcha(driver, store),
	}
}

func (l *GetCaptchaLogic) GetCaptcha() (resp *types.CaptchaResp, err error) {
	// 生成验证码
	id, b64s, _, err := l.captcha.Generate()
	if err != nil {
		return nil, err
	}

	resp = &types.CaptchaResp{
		CaptchaID:    id,
		CaptchaImage: b64s,
		ExpiresIn:    DefaultCaptchaExpire * 60, // seconds
	}
	return resp, nil
}
