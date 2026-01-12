package logic_test

import (
	"context"
	"testing"

	"auth-service/internal/config"
	"auth-service/internal/logic"
	"auth-service/internal/svc"
	"auth-service/internal/types"

	"github.com/mojocn/base64Captcha"
	"github.com/zeromicro/go-zero/core/logx"
)

// MockCaptchaStore implement base64Captcha.Store for testing
type MockCaptchaStore struct {
	store map[string]string
}

func NewMockCaptchaStore() *MockCaptchaStore {
	return &MockCaptchaStore{
		store: make(map[string]string),
	}
}

func (s *MockCaptchaStore) Set(id string, value string) error {
	s.store[id] = value
	return nil
}

func (s *MockCaptchaStore) Get(id string, clear bool) string {
	val := s.store[id]
	if clear {
		delete(s.store, id)
	}
	return val
}

func (s *MockCaptchaStore) Verify(id, answer string, clear bool) bool {
	val := s.Get(id, clear)
	return val == answer
}

// Ensure the helper context setup allows modifying JWT and Captcha
func TestGetCaptchaLogic_GetCaptcha(t *testing.T) {
	// Initialize context (using nil DB as we don't need it here)
	cfg := config.Config{
		Captcha: struct {
			Enable      bool
			ExpiresIn   int64
			Length      int
			CachePrefix string
		}{
			Enable:    true,
			ExpiresIn: 300,
			Length:    6,
		},
	}
	svcCtx := &svc.ServiceContext{
		Config: cfg,
	}

	// Setup valid captcha
	driver := base64Captcha.NewDriverDigit(80, 240, 6, 0.7, 80)
	store := NewMockCaptchaStore()
	svcCtx.Captcha = base64Captcha.NewCaptcha(driver, store)

	l := logic.NewGetCaptchaLogic(context.Background(), svcCtx)
	resp, err := l.GetCaptcha()

	if err != nil {
		t.Fatalf("GetCaptcha failed: %v", err)
	}

	if resp.CaptchaID == "" {
		t.Error("Expected CaptchaID")
	}
	if resp.CaptchaImage == "" {
		t.Error("Expected CaptchaImage")
	}
}

func TestForgotPasswordLogic_ForgotPassword(t *testing.T) {
	svcCtx := &svc.ServiceContext{} // minimal context
	l := logic.NewForgotPasswordLogic(context.Background(), svcCtx)

	// Currently empty implementation
	_, err := l.ForgotPassword(&types.ForgotPasswordReq{
		Email: "test@example.com",
	})

	if err != nil {
		// If implementation is empty, it might return nil error or panic?
		// The scaffolded code returns (resp, err) with named returns, initialized to nil.
		// So err should be nil.
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRefreshLogic_Refresh(t *testing.T) {
	// Setup JWT without Redis (for refresh without blacklist check)
	cfg := config.Config{
		Auth: struct {
			AccessSecret         string
			AccessExpiresIn      int64
			RefreshSecret        string
			RefreshExpiresIn     int64
			BlacklistCachePrefix string
		}{
			AccessSecret:     "secret",
			AccessExpiresIn:  3600,
			RefreshSecret:    "secret",
			RefreshExpiresIn: 7200,
		},
	}

	// Passing nil for redis client to svc.NewJWT causes it to skip blacklist checks
	jwtService := svc.NewJWT(cfg, nil)
	svcCtx := &svc.ServiceContext{
		Config: cfg,
		JWT:    jwtService,
	}

	// Generate a valid token pair
	tokenPair, err := jwtService.Generate(1, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Sleep briefly to ensure timestamps might differ if precision is high,
	// though for JWT usually second-level precision.
	// Logic just checks signature and expiration.

	l := logic.NewRefreshLogic(context.Background(), svcCtx)
	resp, err := l.Refresh(&types.RefreshReq{
		RefreshToken: tokenPair.RefreshToken,
	})

	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("Expected new AccessToken")
	}
	if resp.RefreshToken == "" {
		t.Error("Expected new RefreshToken")
	}
}

func TestLogoutLogic_Logout_Fail(t *testing.T) {
	// Test failure path when Redis is missing (or JWT not capable of blacklist)
	cfg := config.Config{
		Auth: struct {
			AccessSecret         string
			AccessExpiresIn      int64
			RefreshSecret        string
			RefreshExpiresIn     int64
			BlacklistCachePrefix string
		}{
			AccessSecret:     "secret",
			AccessExpiresIn:  3600,
			RefreshSecret:    "secret",
			RefreshExpiresIn: 7200,
		},
	}

	// nil DB -> Logout returns error "rdb is not initialized"
	jwtService := svc.NewJWT(cfg, nil)
	svcCtx := &svc.ServiceContext{
		Config: cfg,
		JWT:    jwtService,
	}

	// We need valid tokens to pass verification steps inside Logout before it hits Redis check
	// Wait, let's check JWT.Logout implementation
	// func (j *JWT) Logout(accessToken string, refreshToken string) error {
	//    if j.rdb == nil { return errors.New("rdb is not initialized") }
	//    ...
	// }
	// So it fails fast. We don't even need valid tokens.

	l := logic.NewLogoutLogic(context.Background(), svcCtx)

	// Turn off error logs for expected error
	logx.Disable()
	defer logx.Disable() // Restore? logx doesn't have simple restore.
	// Actually, we can just ignore the log output.

	req := &types.LogoutReq{
		AccessToken:  "any",
		RefreshToken: "any",
	}

	_, err := l.Logout(req)

	if err == nil {
		t.Error("Expected error because Redis is not initialized")
	}

	// Check coverage: logic file should be covered for the error path
}
