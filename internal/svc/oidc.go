package svc

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"auth-service/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
)

// OIDCProvider OpenID Connect 提供者
type OIDCProvider struct {
	config     config.OIDCConfig
	discovery  *OIDCDiscovery
	httpClient *http.Client
}

// OIDCDiscovery OIDC 发现文档
type OIDCDiscovery struct {
	Issuer                 string   `json:"issuer"`
	AuthorizationEndpoint  string   `json:"authorization_endpoint"`
	TokenEndpoint          string   `json:"token_endpoint"`
	UserInfoEndpoint       string   `json:"userinfo_endpoint"`
	JWKSUri                string   `json:"jwks_uri"`
	ScopesSupported        []string `json:"scopes_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
	EndSessionEndpoint     string   `json:"end_session_endpoint,omitempty"`
}

// OIDCTokenResponse OIDC Token 响应
type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// OIDCUserInfo OIDC 用户信息
type OIDCUserInfo struct {
	Sub               string `json:"sub"`
	Name              string `json:"name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	Picture           string `json:"picture,omitempty"`
	Locale            string `json:"locale,omitempty"`
}

// OIDCState OIDC 状态信息 (用于防止 CSRF)
type OIDCState struct {
	Nonce       string `json:"nonce"`
	RedirectURL string `json:"redirect_url,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

// NewOIDCProvider 创建 OIDC 提供者
func NewOIDCProvider(cfg config.OIDCConfig) (*OIDCProvider, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
	}

	provider := &OIDCProvider{
		config: cfg,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}

	// 获取 OIDC 发现文档
	if err := provider.fetchDiscovery(); err != nil {
		return nil, fmt.Errorf("failed to fetch OIDC discovery: %w", err)
	}

	return provider, nil
}

// fetchDiscovery 获取 OIDC 发现文档
func (p *OIDCProvider) fetchDiscovery() error {
	discoveryURL := strings.TrimSuffix(p.config.ProviderURL, "/") + "/.well-known/openid-configuration"

	resp, err := p.httpClient.Get(discoveryURL)
	if err != nil {
		return fmt.Errorf("failed to fetch discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discovery endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var discovery OIDCDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return fmt.Errorf("failed to decode discovery document: %w", err)
	}

	p.discovery = &discovery
	logx.Infof("OIDC discovery loaded from %s, issuer: %s", discoveryURL, discovery.Issuer)
	return nil
}

// GetAuthorizationURL 获取授权 URL
func (p *OIDCProvider) GetAuthorizationURL(state, nonce string) string {
	params := url.Values{}
	params.Set("client_id", p.config.ClientID)
	params.Set("redirect_uri", p.config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(p.config.Scopes, " "))
	params.Set("state", state)
	params.Set("nonce", nonce)

	return p.discovery.AuthorizationEndpoint + "?" + params.Encode()
}

// ExchangeCode 用授权码交换令牌
func (p *OIDCProvider) ExchangeCode(ctx context.Context, code string) (*OIDCTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", p.config.RedirectURL)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.discovery.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp OIDCTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo 获取用户信息
func (p *OIDCProvider) GetUserInfo(ctx context.Context, accessToken string) (*OIDCUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.discovery.UserInfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var userInfo OIDCUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo: %w", err)
	}

	return &userInfo, nil
}

// RefreshAccessToken 刷新访问令牌
func (p *OIDCProvider) RefreshAccessToken(ctx context.Context, refreshToken string) (*OIDCTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.discovery.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp OIDCTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// GetLogoutURL 获取登出 URL
func (p *OIDCProvider) GetLogoutURL(idToken, postLogoutRedirectURI string) string {
	if p.discovery.EndSessionEndpoint == "" {
		return ""
	}

	params := url.Values{}
	if idToken != "" {
		params.Set("id_token_hint", idToken)
	}
	if postLogoutRedirectURI != "" {
		params.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	}

	return p.discovery.EndSessionEndpoint + "?" + params.Encode()
}

// GetDiscovery 获取发现文档
func (p *OIDCProvider) GetDiscovery() *OIDCDiscovery {
	return p.discovery
}

// IsEnabled 检查是否启用
func (p *OIDCProvider) IsEnabled() bool {
	return p != nil && p.config.Enabled
}
