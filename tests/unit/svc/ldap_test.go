package svc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth-service/internal/logic"
	"auth-service/internal/svc"
	"auth-service/internal/types"
	model "auth-service/model/mysql"

	"github.com/DATA-DOG/go-sqlmock"

	"auth-service/tests/common")

func TestLDAPLogin(t *testing.T) {
	h := common.NewTestHelper(t)
	// Use existing helper but override LDAP
	svcCtx := h.SetupServiceContext(false)

	// Setup Mock LDAP
	mockLDAP := &common.MockLDAPClient{}
	svcCtx.LDAP = mockLDAP

	// Initialize JWT manually for testing
	// Use an anonymous struct compatible with Config.Auth if needed, or just set if defined
	// svcCtx.Config logic in helper.go defines cfg.Auth only when withRedis is true?
	// helper.go defines cfg.Auth inside the if withRedis block as a locally scoped struct literal assign to cfg.Auth?
	// No, Config struct is defined in `internal/config`.
	// helper.go initializes cfg (local var) then assigns fields.

	// We need to set Auth config fields.
	// Since we can't easily modify Config struct in test if it's not exported or complex,
	// let's just create a new JWT with manual config values.

	// Construct a config.Config equivalent logic
	// Using a simpler approach: define a dummy config for NewJWT

	// Is config.Config available? Yes.
	// Does it have Auth field? Yes.

	// Let's rely on internal/config/config.go definition.
	// Start by importing config package

	// Setup JWT with dummy values
	// We can't change svcCtx.Config easily if we don't know the full struct.
	// valid config.Config

	// Re-check helper.go to see how it constructs Config.
	// It seems helper.go constructs Config inline?
	// cfg := config.Config{...}

	// So in ldap_test.go:
	svcCtx.Config.Auth.AccessSecret = "testsecret"
	svcCtx.Config.Auth.AccessExpiresIn = 3600
	svcCtx.Config.Auth.RefreshSecret = "testrefresh"
	svcCtx.Config.Auth.RefreshExpiresIn = 7200

	svcCtx.JWT = svc.NewJWT(svcCtx.Config, nil)

	l := logic.NewLDAPLoginLogic(context.Background(), svcCtx)

	columns := []string{
		"id", "public_id", "nickname", "username", "email", "email_verified",
		"phone", "phone_verified", "password_hash", "password_salt",
		"mfa_secret", "mfa_enabled", "account_status", "failed_login_attempts",
		"lockout_until", "last_login_at", "created_at", "updated_at", "deleted_at",
	}

	t.Run("LDAP Disabled", func(t *testing.T) {
		mockLDAP.IsEnabledFunc = func() bool { return false }

		req := &types.LDAPLoginReq{
			Username: "testuser",
			Password: "password",
		}

		resp, err := l.LDAPLogin(req)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if resp.Code != 1001 {
			t.Errorf("expected code 1001, got %d", resp.Code)
		}
	})

	t.Run("LDAP Auth Failed", func(t *testing.T) {
		mockLDAP.IsEnabledFunc = func() bool { return true }
		mockLDAP.AuthenticateFunc = func(ctx context.Context, u, p string) (*svc.LDAPUserInfo, error) {
			return nil, errors.New("invalid credentials")
		}

		req := &types.LDAPLoginReq{
			Username: "testuser",
			Password: "wrongpassword",
		}

		resp, err := l.LDAPLogin(req)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if resp.Code != 1002 {
			t.Errorf("expected code 1002, got %d", resp.Code)
		}
	})

	t.Run("LDAP Success - Existing User", func(t *testing.T) {
		mockLDAP.IsEnabledFunc = func() bool { return true }
		mockLDAP.AuthenticateFunc = func(ctx context.Context, u, p string) (*svc.LDAPUserInfo, error) {
			return &svc.LDAPUserInfo{
				Username:    "testuser",
				Email:       "test@example.com",
				DisplayName: "Test User",
			}, nil
		}

		// Mock DB for FindOneByUsername
		rows := sqlmock.NewRows(columns).
			AddRow(1, "pub_id_1", "Test User", "testuser", "test@example.com", 1,
				"", 0, "hash", "", "", 0, 1, 0, time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{}) // Fill null/empty for others

		// Use lenient regex for SQL matching to handle backticks
		h.GetMock().ExpectQuery("(?i)select.+from.+user.+where.+username.+").
			WithArgs("testuser").
			WillReturnRows(rows)

		req := &types.LDAPLoginReq{
			Username: "testuser",
			Password: "password",
		}

		resp, err := l.LDAPLogin(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Code != 0 {
			t.Errorf("expected code 0, got %d", resp.Code)
		}

		data := resp.Data.(types.LDAPLoginResp)
		if data.Username != "testuser" {
			t.Errorf("expected username testuser, got %s", data.Username)
		}
		if data.IsNewUser {
			t.Errorf("expected existing user")
		}
	})

	t.Run("LDAP Success - New User", func(t *testing.T) {
		mockLDAP.IsEnabledFunc = func() bool { return true }
		mockLDAP.AuthenticateFunc = func(ctx context.Context, u, p string) (*svc.LDAPUserInfo, error) {
			return &svc.LDAPUserInfo{
				Username:    "newuser",
				Email:       "new@example.com",
				DisplayName: "New User",
			}, nil
		}

		// Mock FindOneByUsername -> Not Found
		h.GetMock().ExpectQuery("(?i)select.+from.+user.+where.+username.+").
			WithArgs("newuser").
			WillReturnError(model.ErrNotFound)

			// Mock FindOneByEmail -> Not Found
		h.GetMock().ExpectQuery("(?i)select.+from.+user.+where.+email.+").
			WithArgs("new@example.com").
			WillReturnError(model.ErrNotFound)

		// Mock FindOneByUsername (Uniqueness Check) -> Not Found
		h.GetMock().ExpectQuery("(?i)select.+from.+user.+where.+username.+").
			WithArgs("newuser").
			WillReturnError(model.ErrNotFound)

		// Mock Insert
		h.GetMock().ExpectExec("(?i)insert.+into.+user.+").
			WillReturnResult(sqlmock.NewResult(2, 1))

			// Mock FindOne (after insert)
		rows := sqlmock.NewRows(columns).
			AddRow(2, "pub_id_2", "New User", "newuser", "new@example.com", 0,
				"", 0, "hash", "", "", 0, 1, 0, time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{})
		h.GetMock().ExpectQuery("(?i)select.+from.+user.+where.+id.+").
			WithArgs(2).
			WillReturnRows(rows)

		req := &types.LDAPLoginReq{
			Username: "newuser",
			Password: "password",
		}

		resp, err := l.LDAPLogin(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Code != 0 {
			t.Errorf("expected code 0, got %d", resp.Code)
		}

		data := resp.Data.(types.LDAPLoginResp)
		if data.Username != "newuser" {
			t.Errorf("expected username newuser, got %s", data.Username)
		}
		if !data.IsNewUser {
			t.Errorf("expected new user")
		}
	})

}
