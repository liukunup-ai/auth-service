package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	Mysql struct {
		DataSource string
	}

	Redis struct {
		Addrs    []string
		DB       int
		Password string
	}

	Auth struct {
		AccessSecret         string
		AccessExpiresIn      int64
		RefreshSecret        string
		RefreshExpiresIn     int64
		BlacklistCachePrefix string
	}

	Captcha struct {
		Enable      bool
		ExpiresIn   int64
		Length      int
		CachePrefix string
	}

	// SSO 配置
	SSO SSOConfig
}

// SSOConfig SSO 统一配置
type SSOConfig struct {
	// 默认身份提供者: local, oidc, ldap
	DefaultProvider string

	// OpenID Connect 配置
	OIDC OIDCConfig

	// LDAP 配置
	LDAP LDAPConfig
}

// OIDCConfig OpenID Connect 配置
type OIDCConfig struct {
	Enabled            bool
	ProviderURL        string // OIDC Provider URL (如 https://accounts.google.com)
	ClientID           string
	ClientSecret       string
	RedirectURL        string   // 回调 URL
	Scopes             []string // 请求的 scopes
	InsecureSkipVerify bool     // 是否跳过 TLS 证书验证
}

// LDAPConfig LDAP 配置
type LDAPConfig struct {
	Enabled         bool
	Host            string   // LDAP 服务器地址
	Port            int      // LDAP 端口 (默认 389, LDAPS 默认 636)
	UseSSL          bool     // 是否使用 SSL
	UseTLS          bool     // 是否使用 StartTLS
	InsecureSkipTLS bool     // 是否跳过 TLS 证书验证 (仅用于测试)
	BindDN          string   // 绑定 DN
	BindPassword    string   // 绑定密码
	BaseDN          string   // 搜索基准 DN
	UserFilter      string   // 用户过滤器 (如 "(uid=%s)" 或 "(sAMAccountName=%s)")
	GroupFilter     string   // 组过滤器 (可选)
	UserAttributes  []string // 需要获取的用户属性
	UsernameAttr    string   // 用户名属性 (如 uid, sAMAccountName)
	EmailAttr       string   // 邮箱属性 (如 mail)
	DisplayNameAttr string   // 显示名称属性 (如 displayName, cn)
	GroupMemberAttr string   // 组成员属性 (如 memberOf)
}
