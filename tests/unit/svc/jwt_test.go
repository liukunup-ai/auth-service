package svc_test

import (
	"context"
	"testing"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/svc"

	"github.com/redis/go-redis/v9"
)

func setupTestJWT(t *testing.T) *svc.JWT {
	// 创建测试用的 Redis 客户端（使用 mock 或者测试 Redis）
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       15, // 使用测试数据库
	})

	// 测试 Redis 连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis is not available, skipping JWT tests")
	}

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(ctx)
		rdb.Close()
	})

	cfg := config.Config{
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

	return svc.NewJWT(cfg, rdb)
}

func TestJWT_Generate(t *testing.T) {
	jwt := setupTestJWT(t)

	userID := uint64(12345)
	username := "testuser"

	tokenPair, err := jwt.Generate(userID, username)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}

	if tokenPair.AccessExpiresAt == 0 {
		t.Error("AccessExpiresAt should not be zero")
	}

	if tokenPair.RefreshExpiresAt == 0 {
		t.Error("RefreshExpiresAt should not be zero")
	}

	if tokenPair.AccessExpiresAt >= tokenPair.RefreshExpiresAt {
		t.Error("AccessToken should expire before RefreshToken")
	}
}

func TestJWT_VerifyAccessToken(t *testing.T) {
	jwt := setupTestJWT(t)

	userID := uint64(12345)
	username := "testuser"

	tokenPair, err := jwt.Generate(userID, username)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	claims, err := jwt.VerifyAccessToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}

	if claims.Username != username {
		t.Errorf("Username = %v, want %v", claims.Username, username)
	}

	if claims.Type != svc.AccessToken {
		t.Errorf("Type = %v, want %v", claims.Type, svc.AccessToken)
	}
}

func TestJWT_VerifyRefreshToken(t *testing.T) {
	jwt := setupTestJWT(t)

	userID := uint64(12345)
	username := "testuser"

	tokenPair, err := jwt.Generate(userID, username)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	claims, err := jwt.VerifyRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("VerifyRefreshToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}

	if claims.Username != username {
		t.Errorf("Username = %v, want %v", claims.Username, username)
	}

	if claims.Type != svc.RefreshToken {
		t.Errorf("Type = %v, want %v", claims.Type, svc.RefreshToken)
	}
}

func TestJWT_VerifyWrongTokenType(t *testing.T) {
	jwt := setupTestJWT(t)

	userID := uint64(12345)
	username := "testuser"

	tokenPair, err := jwt.Generate(userID, username)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 使用 AccessToken 验证 RefreshToken 应该失败
	_, err = jwt.VerifyRefreshToken(tokenPair.AccessToken)
	if err == nil {
		t.Error("VerifyRefreshToken() with AccessToken should fail")
	}

	// 使用 RefreshToken 验证 AccessToken 应该失败
	_, err = jwt.VerifyAccessToken(tokenPair.RefreshToken)
	if err == nil {
		t.Error("VerifyAccessToken() with RefreshToken should fail")
	}
}

func TestJWT_Refresh(t *testing.T) {
	jwt := setupTestJWT(t)

	userID := uint64(12345)
	username := "testuser"

	// 生成初始令牌对
	oldTokenPair, err := jwt.Generate(userID, username)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 等待一小段时间确保时间戳不同
	time.Sleep(time.Second)

	// 刷新令牌
	newTokenPair, err := jwt.Refresh(oldTokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	// 新令牌不应该与旧令牌相同
	if newTokenPair.AccessToken == oldTokenPair.AccessToken {
		t.Error("New AccessToken should be different from old one")
	}

	if newTokenPair.RefreshToken == oldTokenPair.RefreshToken {
		t.Error("New RefreshToken should be different from old one")
	}

	// 新令牌应该可以验证
	claims, err := jwt.VerifyAccessToken(newTokenPair.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}

	// 旧的刷新令牌应该被加入黑名单（再次使用应该失败）
	_, err = jwt.Refresh(oldTokenPair.RefreshToken)
	if err == nil {
		t.Error("Using old RefreshToken again should fail")
	}
}

func TestJWT_Logout(t *testing.T) {
	jwt := setupTestJWT(t)

	userID := uint64(12345)
	username := "testuser"

	tokenPair, err := jwt.Generate(userID, username)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 验证令牌有效
	_, err = jwt.VerifyAccessToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccessToken() before logout error = %v", err)
	}

	// 登出
	err = jwt.Logout(tokenPair.AccessToken, tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	// 登出后令牌应该无效
	_, err = jwt.VerifyAccessToken(tokenPair.AccessToken)
	if err == nil {
		t.Error("AccessToken should be invalid after logout")
	}

	_, err = jwt.VerifyRefreshToken(tokenPair.RefreshToken)
	if err == nil {
		t.Error("RefreshToken should be invalid after logout")
	}
}

func TestJWT_BlacklistToken(t *testing.T) {
	jwt := setupTestJWT(t)

	userID := uint64(12345)
	username := "testuser"

	tokenPair, err := jwt.Generate(userID, username)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 将令牌加入黑名单
	err = jwt.BlacklistToken(tokenPair.AccessToken, time.Hour)
	if err != nil {
		t.Fatalf("BlacklistToken() error = %v", err)
	}

	// 验证应该失败
	_, err = jwt.VerifyAccessToken(tokenPair.AccessToken)
	if err == nil {
		t.Error("Token should be invalid after being blacklisted")
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	jwt := setupTestJWT(t)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "空令牌",
			token: "",
		},
		{
			name:  "无效格式",
			token: "invalid-token",
		},
		{
			name:  "伪造的令牌",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := jwt.VerifyAccessToken(tt.token)
			if err == nil {
				t.Errorf("VerifyAccessToken() should fail for %s", tt.name)
			}
		})
	}
}
