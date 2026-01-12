package logic

import (
	"context"

	"auth-service/internal/svc"
	"auth-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SSOProvidersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSSOProvidersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SSOProvidersLogic {
	return &SSOProvidersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SSOProviders 获取可用的 SSO 提供者列表
func (l *SSOProvidersLogic) SSOProviders() (*types.SSOProvidersResp, error) {
	providers := []types.SSOProvider{
		{
			ID:       string(types.SSOProviderLocal),
			Name:     "本地账号",
			Type:     string(types.SSOProviderLocal),
			Enabled:  true,
			LoginURL: "/api/v1/login",
		},
	}

	// 检查 OIDC
	if l.svcCtx.OIDC.IsEnabled() {
		providers = append(providers, types.SSOProvider{
			ID:       string(types.SSOProviderOIDC),
			Name:     "OpenID Connect",
			Type:     string(types.SSOProviderOIDC),
			Enabled:  true,
			LoginURL: "/api/v1/sso/oidc/login",
		})
	}

	// 检查 LDAP
	if l.svcCtx.LDAP.IsEnabled() {
		providers = append(providers, types.SSOProvider{
			ID:       string(types.SSOProviderLDAP),
			Name:     "LDAP / Active Directory",
			Type:     string(types.SSOProviderLDAP),
			Enabled:  true,
			LoginURL: "/api/v1/sso/ldap/login",
		})
	}

	defaultProvider := l.svcCtx.Config.SSO.DefaultProvider
	if defaultProvider == "" {
		defaultProvider = string(types.SSOProviderLocal)
	}

	return &types.SSOProvidersResp{
		DefaultProvider: defaultProvider,
		Providers:       providers,
	}, nil
}
