package svc_test

import (
	"context"
	"testing"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/svc"

	"github.com/golang-jwt/jwt"
)

func TestJwtClaimsAdapter(t *testing.T) {
	// Future expiration
	future := time.Now().Add(time.Hour).Unix()
	claims := &svc.CustomClaims{
		UserID: 123,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: future,
		},
	}
	adapter := &svc.JwtClaimsAdapter{
		Claims: claims,
	}

	if adapter.GetUserID() != 123 {
		t.Errorf("Expected UserID 123, got %d", adapter.GetUserID())
	}
	if !adapter.VerifyExpiresAt() {
		t.Error("Expected valid token")
	}

	// Past expiration
	past := time.Now().Add(-time.Hour).Unix()
	claims.StandardClaims.ExpiresAt = past
	if adapter.VerifyExpiresAt() {
		t.Error("Expected expired token")
	}
}

func TestNewProviders_Disabled(t *testing.T) {
	// OIDC Disabled
	oidcCfg := config.OIDCConfig{Enabled: false}
	p, err := svc.NewOIDCProvider(oidcCfg)
	if err != nil {
		t.Errorf("NewOIDCProvider error: %v", err)
	}
	if p != nil {
		t.Error("Expected nil provider when disabled")
	}

	// LDAP Disabled
	ldapCfg := config.LDAPConfig{Enabled: false}
	l, err := svc.NewLDAPProvider(ldapCfg)
	if err != nil {
		t.Errorf("NewLDAPProvider error: %v", err)
	}
	if l != nil {
		t.Error("Expected nil provider when disabled")
	}
}

func TestRedisStore_New(t *testing.T) {
	// Just test instantiation
	// We pass nil as client, it will panic if we call methods, but NewRedisStore itself shouldn't (if implemented correct).
	// Signature: ctx, client, prefix, ttl
	rs := svc.NewRedisStore(context.Background(), nil, "prefix:", time.Minute*5)
	if rs == nil {
		t.Error("NewRedisStore returned nil")
	}
}
