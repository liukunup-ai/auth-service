package logic_test

import (
	"auth-service/internal/logic"
	"auth-service/internal/svc"
	"context"
	"testing"

	"auth-service/tests/common")

func TestHealthCheckLogic_HealthCheck(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	l := logic.NewHealthCheckLogic(context.Background(), svcCtx)
	resp, err := l.HealthCheck()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.Code != 200 {
		t.Errorf("Expected code 200, got %d", resp.Code)
	}
	if resp.Message != "OK" {
		t.Errorf("Expected message OK, got %s", resp.Message)
	}
}

func TestConfirmPasswordLogic_ConfirmPassword(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	l := logic.NewConfirmPasswordLogic(context.Background(), svcCtx)
	// Currently placeholder implementation
	resp, err := l.ConfirmPassword(nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.Code != 200 {
		t.Errorf("Expected code 200, got %d", resp.Code)
	}
}

func TestSSOProvidersLogic_Get(t *testing.T) {
	// Mock implementation
	mockOIDC := &common.MockOIDCClient{
		IsEnabledFunc: func() bool { return true },
	}
	mockLDAP := &common.MockLDAPClient{
		IsEnabledFunc: func() bool { return true },
	}

	svcCtx := &svc.ServiceContext{
		OIDC: mockOIDC,
		LDAP: mockLDAP,
	}

	l := logic.NewSSOProvidersLogic(context.Background(), svcCtx)
	resp, err := l.SSOProviders()

	if err != nil {
		t.Fatalf("SSOProviders failed: %v", err)
	}

	foundOIDC := false
	foundLDAP := false
	for _, p := range resp.Providers {
		if p.Name == "OpenID Connect" {
			foundOIDC = true
		}
		if p.Name == "LDAP / Active Directory" {
			foundLDAP = true
		}
	}

	if !foundOIDC {
		t.Error("Expected OIDC provider")
	}
	if !foundLDAP {
		t.Error("Expected LDAP provider")
	}
}
