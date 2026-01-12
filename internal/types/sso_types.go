package types

// ===================== SSO 相关类型定义 =====================

// SSOProviderType SSO 提供者类型
type SSOProviderType string

const (
	SSOProviderLocal SSOProviderType = "local" // 本地认证
	SSOProviderOIDC  SSOProviderType = "oidc"  // OpenID Connect
	SSOProviderLDAP  SSOProviderType = "ldap"  // LDAP
)

// ===================== OpenID Connect 类型 =====================

// OIDCLoginReq is defined in types.go

// OIDCLoginResp OIDC 登录响应
type OIDCLoginResp struct {
	AuthorizationURL string `json:"authorizationUrl"` // 跳转到 IDP 的授权 URL
	State            string `json:"state"`            // CSRF 防护 state
}

// OIDCCallbackReq is defined in types.go

// OIDCCallbackResp OIDC 回调响应 (登录成功)
type OIDCCallbackResp struct {
	UserID           string `json:"userId"`
	Username         string `json:"username"`
	Email            string `json:"email,optional"`
	AccessToken      string `json:"accessToken"`
	AccessExpiresAt  int64  `json:"accessExpiresAt"`
	RefreshToken     string `json:"refreshToken"`
	RefreshExpiresAt int64  `json:"refreshExpiresAt"`
	TokenType        string `json:"tokenType" default:"Bearer"`
	IsNewUser        bool   `json:"isNewUser"` // 是否为新用户 (首次 SSO 登录)
	Provider         string `json:"provider"`  // SSO 提供者
}

// ===================== LDAP 类型 =====================

// LDAPLoginReq is defined in types.go

// LDAPLoginResp LDAP 登录响应
type LDAPLoginResp struct {
	UserID           string   `json:"userId"`
	Username         string   `json:"username"`
	Email            string   `json:"email,optional"`
	DisplayName      string   `json:"displayName,optional"`
	Groups           []string `json:"groups,optional"`
	AccessToken      string   `json:"accessToken"`
	AccessExpiresAt  int64    `json:"accessExpiresAt"`
	RefreshToken     string   `json:"refreshToken"`
	RefreshExpiresAt int64    `json:"refreshExpiresAt"`
	TokenType        string   `json:"tokenType" default:"Bearer"`
	IsNewUser        bool     `json:"isNewUser"`
	Provider         string   `json:"provider"`
}

// ===================== SSO 统一类型 =====================

// SSOProvidersResp is defined in types.go

// SSOUserInfo SSO 用户信息 (统一格式)
type SSOUserInfo struct {
	Provider       string            `json:"provider"`       // SSO 提供者
	ProviderUserID string            `json:"providerUserId"` // 提供者侧的用户 ID
	Email          string            `json:"email,optional"`
	Username       string            `json:"username,optional"`
	DisplayName    string            `json:"displayName,optional"`
	FirstName      string            `json:"firstName,optional"`
	LastName       string            `json:"lastName,optional"`
	Picture        string            `json:"picture,optional"`    // 头像 URL
	Groups         []string          `json:"groups,optional"`     // 用户组
	Attributes     map[string]string `json:"attributes,optional"` // 额外属性
	EmailVerified  bool              `json:"emailVerified"`
}

// SSOLinkReq 关联 SSO 账号请求
type SSOLinkReq struct {
	Provider string `json:"provider" validate:"required,oneof=oidc ldap"`
}

// SSOLinkResp 关联 SSO 账号响应
type SSOLinkResp struct {
	AuthorizationURL string `json:"authorizationUrl,optional"` // OIDC 需要跳转
	Message          string `json:"message,optional"`
}

// SSOUnlinkReq 解除 SSO 账号关联请求
type SSOUnlinkReq struct {
	Provider string `json:"provider" validate:"required,oneof=oidc ldap"`
}

// SSOLinkedAccountsResp 已关联的 SSO 账号响应
type SSOLinkedAccountsResp struct {
	Accounts []SSOLinkedAccount `json:"accounts"`
}

// SSOLinkedAccount 已关联的 SSO 账号信息
type SSOLinkedAccount struct {
	Provider       string `json:"provider"`
	ProviderUserID string `json:"providerUserId"`
	Email          string `json:"email,optional"`
	LinkedAt       int64  `json:"linkedAt"` // Unix 时间戳
}
