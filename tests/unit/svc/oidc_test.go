package svc_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/logic"
	"auth-service/internal/svc"
	"auth-service/internal/types"
	model "auth-service/model/mysql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/conf"

	"auth-service/tests/common"
)

func TestOIDC(t *testing.T) {
	// Load actual configuration
	var c config.Config
	// Get current file path to locate the config file
	_, filename, _, _ := runtime.Caller(0)
	configFile := filepath.Join(filepath.Dir(filename), "../../../etc/auth-api.yaml")
	if err := conf.Load(configFile, &c); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	h := common.NewTestHelper(t)
	// Setup Service Context with Mock Redis
	svcCtx := h.SetupServiceContext(true)
	// Override config with loaded config
	svcCtx.Config = c

	// Check if Redis is actually available, if not, we can't test OIDC flow fully unless we mock Redis
	// But common.TestHelper skips if Redis is down.

	mockOIDC := &common.MockOIDCClient{}
	svcCtx.OIDC = mockOIDC

	loginLogic := logic.NewOIDCLoginLogic(context.Background(), svcCtx)
	callbackLogic := logic.NewOIDCCallbackLogic(context.Background(), svcCtx)

	t.Run("Login URL Generation", func(t *testing.T) {
		mockOIDC.IsEnabledFunc = func() bool { return c.SSO.OIDC.Enabled }
		mockOIDC.GetAuthorizationURLFunc = func(state, nonce string) string {
			// Simulate generating URL using real config values
			// In a real implementation this would use c.SSO.OIDC.ProviderURL, ClientID, etc.
			// For this test, we accept what the logic calls, but verify the logic *would* redirect.
			// Note: The logic itself calls OIDC.GetAuthorizationURL.
			// We should return a URL that looks like the real provider's URL to confirm context.
			return c.SSO.OIDC.ProviderURL + "/protocol/openid-connect/auth?state=" + state
		}

		req := &types.OIDCLoginReq{
			RedirectURL: c.SSO.OIDC.RedirectURL,
		}

		resp, err := loginLogic.OIDCLogin(req)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if resp.Code != 0 {
			t.Errorf("expected code 0, got %d", resp.Code)
		}

		data := resp.Data.(types.OIDCLoginResp)
		if data.AuthorizationURL == "" {
			t.Errorf("expected auth url")
		}

		// If Redis is working, verify state
		ctx := context.Background()
		if svcCtx.Redis != nil {
			val, err := svcCtx.Redis.Get(ctx, "auth:oidc:state:"+data.State).Result()
			// The logic stores the *original* redirect URL (req.RedirectURL) in Redis
			if err == nil && val != c.SSO.OIDC.RedirectURL {
				t.Errorf("expected redirect url in redis to be %s, got %s", c.SSO.OIDC.RedirectURL, val)
			}
		}
	})

	t.Run("Callback Success", func(t *testing.T) {
		mockOIDC.IsEnabledFunc = func() bool { return true }
		state := "test-state"

		// Setup Redis State
		if svcCtx.Redis != nil {
			svcCtx.Redis.Set(context.Background(), "auth:oidc:state:"+state, "/", time.Minute)
		} else {
			t.Skip("Redis not available")
		}

		mockOIDC.ExchangeCodeFunc = func(ctx context.Context, code string) (*svc.OIDCTokenResponse, error) {
			return &svc.OIDCTokenResponse{
				AccessToken: "access-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}, nil
		}

		mockOIDC.GetUserInfoFunc = func(ctx context.Context, token string) (*svc.OIDCUserInfo, error) {
			return &svc.OIDCUserInfo{
				Sub:               "oidc-sub",
				Email:             "oidc@example.com",
				PreferredUsername: "oidcuser",
			}, nil
		}

		// Mock DB Find User (Assume New User)
		// 1. FindOneByEmail -> NotFound
		h.GetMock().ExpectQuery("(?i)select.+from.+user.+where.+email.+").
			WithArgs("oidc@example.com").
			WillReturnError(model.ErrNotFound)

		// 2. FindOneByUsername -> NotFound
		h.GetMock().ExpectQuery("(?i)select.+from.+user.+where.+username.+").
			WithArgs("oidcuser").
			WillReturnError(model.ErrNotFound)

		// 3. Insert
		h.GetMock().ExpectExec("(?i)insert.+into.+user.+").
			WillReturnResult(sqlmock.NewResult(3, 1))

		// 4. FindOne
		rows := sqlmock.NewRows([]string{
			"id", "public_id", "nickname", "username", "email", "email_verified",
			"phone", "phone_verified", "password_hash", "password_salt", "mfa_secret",
			"mfa_enabled", "account_status", "failed_login_attempts", "lockout_until",
			"last_login_at", "created_at", "updated_at", "deleted_at",
		}).AddRow(
			3, "pub_id_3", sql.NullString{}, "oidcuser", "oidc@example.com", 0,
			sql.NullString{}, 0, "hash", sql.NullString{}, sql.NullString{},
			0, 1, 0, sql.NullTime{}, sql.NullTime{}, time.Now(), time.Now(), sql.NullTime{},
		)
		h.GetMock().ExpectQuery("(?i)select.+from.+user.+where.+id.+").
			WithArgs(3).
			WillReturnRows(rows)

		req := &types.OIDCCallbackReq{
			State: state,
			Code:  "auth-code",
		}

		resp, err := callbackLogic.OIDCCallback(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Code != 0 {
			t.Errorf("expected code 0, got %d", resp.Code)
		}

		data := resp.Data.(types.OIDCCallbackResp)
		if data.Username != "oidcuser" {
			t.Errorf("expected oidcuser, got %s", data.Username)
		}
	})

	t.Run("Callback Success - Test User", func(t *testing.T) {
		mockOIDC.IsEnabledFunc = func() bool { return true }
		state := "test-state-user"
		// Using the credentials provided by user: testuser
		// Password '123456@ABC' is conceptually associated with this user
		testEmail := "testuser@example.com"
		testUsername := "testuser"

		// Setup Redis State
		if svcCtx.Redis != nil {
			svcCtx.Redis.Set(context.Background(), "auth:oidc:state:"+state, "/", time.Minute)
		} else {
			t.Skip("Redis not available")
		}

		mockOIDC.ExchangeCodeFunc = func(ctx context.Context, code string) (*svc.OIDCTokenResponse, error) {
			return &svc.OIDCTokenResponse{
				AccessToken: "access-token-testuser",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}, nil
		}

		mockOIDC.GetUserInfoFunc = func(ctx context.Context, token string) (*svc.OIDCUserInfo, error) {
			return &svc.OIDCUserInfo{
				Sub:               "oidc-sub-testuser",
				Email:             testEmail,
				PreferredUsername: testUsername,
			}, nil
		}

		// Mock DB Find User (Existing User Scenario)
		// 1. FindOneByEmail -> Found
		rows := sqlmock.NewRows([]string{
			"id", "public_id", "nickname", "username", "email", "email_verified",
			"phone", "phone_verified", "password_hash", "password_salt", "mfa_secret",
			"mfa_enabled", "account_status", "failed_login_attempts", "lockout_until",
			"last_login_at", "created_at", "updated_at", "deleted_at",
		}).AddRow(
			10, "pub_id_10", sql.NullString{}, testUsername, testEmail, 0,
			sql.NullString{}, 0, "hash_of_123456@ABC", sql.NullString{}, sql.NullString{},
			0, 1, 0, sql.NullTime{}, sql.NullTime{}, time.Now(), time.Now(), sql.NullTime{},
		)

		h.GetMock().ExpectQuery("(?i)select.+from.+user.+where.+email.+").
			WithArgs(testEmail).
			WillReturnRows(rows)

		req := &types.OIDCCallbackReq{
			State: state,
			Code:  "auth-code-testuser",
		}

		resp, err := callbackLogic.OIDCCallback(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Code != 0 {
			t.Errorf("expected code 0, got %d", resp.Code)
		}

		data := resp.Data.(types.OIDCCallbackResp)
		if data.Username != testUsername {
			t.Errorf("expected %s, got %s", testUsername, data.Username)
		}
	})
}
