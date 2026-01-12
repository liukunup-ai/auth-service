package common

import (
	"auth-service/internal/svc"
	"context"
)

// Manual Mock for LDAPClient
type MockLDAPClient struct {
	AuthenticateFunc  func(ctx context.Context, username, password string) (*svc.LDAPUserInfo, error)
	GetUserGroupsFunc func(ctx context.Context, userDN string) ([]string, error)
	IsEnabledFunc     func() bool
}

func (m *MockLDAPClient) Authenticate(ctx context.Context, username, password string) (*svc.LDAPUserInfo, error) {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(ctx, username, password)
	}
	return nil, nil
}

func (m *MockLDAPClient) GetUserGroups(ctx context.Context, userDN string) ([]string, error) {
	if m.GetUserGroupsFunc != nil {
		return m.GetUserGroupsFunc(ctx, userDN)
	}
	return nil, nil
}

func (m *MockLDAPClient) IsEnabled() bool {
	if m.IsEnabledFunc != nil {
		return m.IsEnabledFunc()
	}
	return false
}

// Manual Mock for OIDCClient
type MockOIDCClient struct {
	IsEnabledFunc           func() bool
	GetAuthorizationURLFunc func(state, nonce string) string
	ExchangeCodeFunc        func(ctx context.Context, code string) (*svc.OIDCTokenResponse, error)
	GetUserInfoFunc         func(ctx context.Context, accessToken string) (*svc.OIDCUserInfo, error)
}

func (m *MockOIDCClient) IsEnabled() bool {
	if m.IsEnabledFunc != nil {
		return m.IsEnabledFunc()
	}
	return false
}

func (m *MockOIDCClient) GetAuthorizationURL(state, nonce string) string {
	if m.GetAuthorizationURLFunc != nil {
		return m.GetAuthorizationURLFunc(state, nonce)
	}
	return ""
}

func (m *MockOIDCClient) ExchangeCode(ctx context.Context, code string) (*svc.OIDCTokenResponse, error) {
	if m.ExchangeCodeFunc != nil {
		return m.ExchangeCodeFunc(ctx, code)
	}
	return nil, nil
}

func (m *MockOIDCClient) GetUserInfo(ctx context.Context, accessToken string) (*svc.OIDCUserInfo, error) {
	if m.GetUserInfoFunc != nil {
		return m.GetUserInfoFunc(ctx, accessToken)
	}
	return nil, nil
}
