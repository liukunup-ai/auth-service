package svc

import (
	"context"
)

// LDAPClient defines the interface for LDAP operations
type LDAPClient interface {
	Authenticate(ctx context.Context, username, password string) (*LDAPUserInfo, error)
	GetUserGroups(ctx context.Context, userDN string) ([]string, error)
	IsEnabled() bool
}

// OIDCClient defines the interface for OIDC operations
type OIDCClient interface {
	IsEnabled() bool
	GetAuthorizationURL(state, nonce string) string
	ExchangeCode(ctx context.Context, code string) (*OIDCTokenResponse, error)
	GetUserInfo(ctx context.Context, accessToken string) (*OIDCUserInfo, error)
}
