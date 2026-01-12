package benchmark_test

import (
	"testing"

	"auth-service/internal/config"
	"auth-service/internal/svc"

	"github.com/redis/go-redis/v9"
)

// BenchmarkPasswordHash 基准测试密码哈希
func BenchmarkPasswordHash(b *testing.B) {
	encoder := &svc.PasswordEncoder{}
	password := "testpassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.Hash(password)
	}
}

// BenchmarkPasswordCompare 基准测试密码验证
func BenchmarkPasswordCompare(b *testing.B) {
	encoder := &svc.PasswordEncoder{}
	password := "testpassword123"
	hash := encoder.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.Compare(hash, password)
	}
}

// BenchmarkJWTGenerate 基准测试 JWT 生成
func BenchmarkJWTGenerate(b *testing.B) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       15,
	})
	defer rdb.Close()

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
			BlacklistCachePrefix: "bench:blacklist:",
		},
	}

	jwt := svc.NewJWT(cfg, rdb)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jwt.Generate(uint64(i), "benchuser")
	}
}

// BenchmarkJWTVerify 基准测试 JWT 验证
func BenchmarkJWTVerify(b *testing.B) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       15,
	})
	defer rdb.Close()

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
			BlacklistCachePrefix: "bench:blacklist:",
		},
	}

	jwt := svc.NewJWT(cfg, rdb)

	// 预先生成一个令牌
	tokenPair, err := jwt.Generate(12345, "benchuser")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jwt.VerifyAccessToken(tokenPair.AccessToken)
	}
}

// BenchmarkPasswordHashParallel 并发基准测试密码哈希
func BenchmarkPasswordHashParallel(b *testing.B) {
	encoder := &svc.PasswordEncoder{}
	password := "testpassword123"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			encoder.Hash(password)
		}
	})
}

// BenchmarkJWTGenerateParallel 并发基准测试 JWT 生成
func BenchmarkJWTGenerateParallel(b *testing.B) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       15,
	})
	defer rdb.Close()

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
			BlacklistCachePrefix: "bench:blacklist:",
		},
	}

	jwt := svc.NewJWT(cfg, rdb)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			jwt.Generate(uint64(i), "benchuser")
			i++
		}
	})
}
