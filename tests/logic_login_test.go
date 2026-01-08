package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/logic"
	"auth-service/internal/svc"
	"auth-service/internal/types"
	model "auth-service/model/mysql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func setupTestServiceContextWithJWT(t *testing.T, mock sqlmock.Sqlmock) *svc.ServiceContext {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	conn := sqlx.NewSqlConnFromDB(db)

	// 创建测试用的 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       15, // 使用测试数据库
	})

	// 测试 Redis 连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis is not available, skipping login tests")
	}

	t.Cleanup(func() {
		rdb.FlushDB(ctx)
		rdb.Close()
	})

	cfg := config.Config{
		Captcha: struct {
			Enable      bool
			ExpiresIn   int64
			Length      int
			CachePrefix string
		}{
			Enable:      false, // 测试时禁用验证码
			ExpiresIn:   300,
			Length:      6,
			CachePrefix: "test:captcha:",
		},
		Auth: struct {
			AccessSecret         string
			AccessExpiresIn      int64
			RefreshSecret        string
			RefreshExpiresIn     int64
			BlacklistCachePrefix string
		}{
			AccessSecret:         "test-access-secret-key-12345678",
			AccessExpiresIn:      3600,
			RefreshSecret:        "test-refresh-secret-key-12345678",
			RefreshExpiresIn:     7200,
			BlacklistCachePrefix: "test:blacklist:",
		},
	}

	jwt := svc.NewJWT(cfg, rdb)

	return &svc.ServiceContext{
		Config:          cfg,
		DB:              conn,
		Redis:           rdb,
		PasswordEncoder: &svc.PasswordEncoder{},
		JWT:             jwt,
	}
}

func TestLoginLogic_Login_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// 允许无序匹配（因为 mr.Finish 并发执行查询）
	mock.MatchExpectationsInOrder(false)

	encoder := &svc.PasswordEncoder{}
	passwordHash := encoder.Hash("password123")

	// 创建测试用户数据（包含所有字段）
	now := time.Now()
	userRows := sqlmock.NewRows([]string{
		"id", "public_id", "nickname", "username", "email", "email_verified",
		"phone", "phone_verified", "password_hash", "password_salt",
		"mfa_secret", "mfa_enabled", "account_status", "failed_login_attempts",
		"lockout_until", "last_login_at", "created_at", "updated_at", "deleted_at",
	}).AddRow(
		1, "user123", nil, "testuser", "test@example.com", 0,
		nil, 0, passwordHash, nil,
		nil, 0, model.UserStatusActive, 0,
		nil, nil, now, now, nil,
	)

	// 期望通过用户名查询（并发查询，需要设置所有期望）
	mock.ExpectQuery("SELECT \\* FROM user WHERE username = \\?").
		WithArgs("testuser").
		WillReturnRows(userRows)

	// email 和 phone 查询返回 not found
	mock.ExpectQuery("SELECT \\* FROM user WHERE email = \\?").
		WithArgs("testuser").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery("SELECT \\* FROM user WHERE phone = \\?").
		WithArgs("testuser").
		WillReturnError(sql.ErrNoRows)

	// 期望更新最后登录时间
	mock.ExpectExec("UPDATE user SET last_login_at = NOW\\(\\)").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	svcCtx := setupTestServiceContextWithJWT(t, mock)
	svcCtx.DB = sqlx.NewSqlConnFromDB(db)

	l := logic.NewLoginLogic(context.Background(), svcCtx)

	req := &types.LoginReq{
		Username: "testuser",
		Password: "password123",
	}

	resp, err := l.Login(req)

	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if resp.UserID != "user123" {
		t.Errorf("UserID = %v, want %v", resp.UserID, "user123")
	}

	if resp.Username != "testuser" {
		t.Errorf("Username = %v, want %v", resp.Username, "testuser")
	}

	if resp.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}

	if resp.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestLoginLogic_Login_InvalidPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// 允许无序匹配（因为 mr.Finish 并发执行查询）
	mock.MatchExpectationsInOrder(false)

	encoder := &svc.PasswordEncoder{}
	passwordHash := encoder.Hash("correctpassword")

	now := time.Now()
	userRows := sqlmock.NewRows([]string{
		"id", "public_id", "nickname", "username", "email", "email_verified",
		"phone", "phone_verified", "password_hash", "password_salt",
		"mfa_secret", "mfa_enabled", "account_status", "failed_login_attempts",
		"lockout_until", "last_login_at", "created_at", "updated_at", "deleted_at",
	}).AddRow(
		1, "user123", nil, "testuser", "test@example.com", 0,
		nil, 0, passwordHash, nil,
		nil, 0, model.UserStatusActive, 0,
		nil, nil, now, now, nil,
	)

	mock.ExpectQuery("SELECT \\* FROM user WHERE username = \\?").
		WithArgs("testuser").
		WillReturnRows(userRows)

	// email 和 phone 查询返回 not found
	mock.ExpectQuery("SELECT \\* FROM user WHERE email = \\?").
		WithArgs("testuser").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery("SELECT \\* FROM user WHERE phone = \\?").
		WithArgs("testuser").
		WillReturnError(sql.ErrNoRows)

	svcCtx := setupTestServiceContextWithJWT(t, mock)
	svcCtx.DB = sqlx.NewSqlConnFromDB(db)

	l := logic.NewLoginLogic(context.Background(), svcCtx)

	req := &types.LoginReq{
		Username: "testuser",
		Password: "wrongpassword",
	}

	_, err = l.Login(req)

	if err != types.ErrInvalidPassword {
		t.Errorf("Expected ErrInvalidPassword, got %v", err)
	}
}

func TestLoginLogic_Login_UserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// 允许无序匹配（因为 mr.Finish 并发执行查询）
	mock.MatchExpectationsInOrder(false)

	// 所有查询都返回 ErrNoRows
	mock.ExpectQuery("SELECT \\* FROM user WHERE username = \\?").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery("SELECT \\* FROM user WHERE email = \\?").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery("SELECT \\* FROM user WHERE phone = \\?").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	svcCtx := setupTestServiceContextWithJWT(t, mock)
	svcCtx.DB = sqlx.NewSqlConnFromDB(db)

	l := logic.NewLoginLogic(context.Background(), svcCtx)

	req := &types.LoginReq{
		Username: "nonexistent",
		Password: "password123",
	}

	_, err = l.Login(req)

	if err != types.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}
