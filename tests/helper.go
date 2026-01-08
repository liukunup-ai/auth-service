package tests

import (
	"context"
	"testing"

	"auth-service/internal/config"
	"auth-service/internal/svc"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// TestHelper 提供测试辅助函数
type TestHelper struct {
	t       *testing.T
	mock    sqlmock.Sqlmock
	db      sqlx.SqlConn
	redis   redis.UniversalClient
	svcCtx  *svc.ServiceContext
	cleanup []func()
}

// NewTestHelper 创建测试辅助对象
func NewTestHelper(t *testing.T) *TestHelper {
	h := &TestHelper{
		t:       t,
		cleanup: make([]func(), 0),
	}

	// 注册清理函数
	t.Cleanup(func() {
		h.Cleanup()
	})

	return h
}

// SetupMockDB 设置模拟数据库
func (h *TestHelper) SetupMockDB() sqlmock.Sqlmock {
	db, mock, err := sqlmock.New()
	if err != nil {
		h.t.Fatalf("Failed to create sqlmock: %v", err)
	}

	h.db = sqlx.NewSqlConnFromDB(db)
	h.mock = mock

	h.cleanup = append(h.cleanup, func() {
		db.Close()
	})

	return mock
}

// SetupTestRedis 设置测试Redis
func (h *TestHelper) SetupTestRedis() redis.UniversalClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456", // Redis 密码
		DB:       15,       // 使用测试数据库
	})

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		h.t.Skip("Redis is not available, skipping test")
	}

	h.redis = rdb

	h.cleanup = append(h.cleanup, func() {
		rdb.FlushDB(ctx)
		rdb.Close()
	})

	return rdb
}

// SetupServiceContext 设置服务上下文
func (h *TestHelper) SetupServiceContext(withRedis bool) *svc.ServiceContext {
	if h.db == nil {
		h.SetupMockDB()
	}

	cfg := config.Config{
		Captcha: struct {
			Enable      bool
			ExpiresIn   int64
			Length      int
			CachePrefix string
		}{
			Enable:      false,
			ExpiresIn:   300,
			Length:      6,
			CachePrefix: "test:captcha:",
		},
	}

	svcCtx := &svc.ServiceContext{
		Config:          cfg,
		DB:              h.db,
		PasswordEncoder: &svc.PasswordEncoder{},
	}

	if withRedis {
		if h.redis == nil {
			h.SetupTestRedis()
		}

		cfg.Auth = struct {
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
		}

		svcCtx.Redis = h.redis
		svcCtx.JWT = svc.NewJWT(cfg, h.redis)
	}

	h.svcCtx = svcCtx
	return svcCtx
}

// GetMock 获取sqlmock对象
func (h *TestHelper) GetMock() sqlmock.Sqlmock {
	return h.mock
}

// GetServiceContext 获取服务上下文
func (h *TestHelper) GetServiceContext() *svc.ServiceContext {
	return h.svcCtx
}

// Cleanup 清理测试资源
func (h *TestHelper) Cleanup() {
	for i := len(h.cleanup) - 1; i >= 0; i-- {
		h.cleanup[i]()
	}
}

// VerifyMockExpectations 验证所有mock期望都被满足
func (h *TestHelper) VerifyMockExpectations() {
	if h.mock != nil {
		if err := h.mock.ExpectationsWereMet(); err != nil {
			h.t.Errorf("Unfulfilled mock expectations: %v", err)
		}
	}
}
