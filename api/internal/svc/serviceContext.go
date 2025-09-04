package svc

import (
	"context"
	"fmt"

	"auth-service/api/internal/config"
	"auth-service/api/internal/middleware"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config          config.Config
	DB              sqlx.SqlConn
	Redis           redis.UniversalClient
	AuthInterceptor rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MySQL 连接
	conn := sqlx.NewMysql(c.Mysql.DataSource)

	// 初始化 Redis 客户端
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    c.Redis.Addrs,
		DB:       c.Redis.DB,
		Password: c.Redis.Password,
	})
	// (可选) 测试联通性
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("failed to connect to redis: %v", err))
	}

	return &ServiceContext{
		Config:          c,
		DB:              conn,
		Redis:           rdb,
		AuthInterceptor: middleware.NewAuthInterceptorMiddleware().Handle,
	}
}
