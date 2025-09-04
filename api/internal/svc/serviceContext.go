package svc

import (
	"context"
	"fmt"
	"time"

	"auth-service/api/internal/config"
	"auth-service/api/internal/middleware"
	mysql "auth-service/model/mysql"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config          config.Config
	DB              sqlx.SqlConn
	Redis           redis.UniversalClient
	AuthInterceptor rest.Middleware
	Captcha         *base64Captcha.Captcha
	PasswordEncoder *PasswordEncoder
	//
	UserModel mysql.UserModel
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

	// 初始化 Captcha 依赖项
	driver := base64Captcha.NewDriverDigit(80, 240, c.Captcha.Length, 0.7, 80)
	store := NewRedisStore(context.Background(), rdb, "captcha:", time.Duration(c.Captcha.Expire)*time.Second)

	return &ServiceContext{
		Config:          c,
		DB:              conn,
		Redis:           rdb,
		AuthInterceptor: middleware.NewAuthInterceptorMiddleware().Handle,
		Captcha:         base64Captcha.NewCaptcha(driver, store),
		PasswordEncoder: &PasswordEncoder{},
		UserModel:       mysql.NewUserModel(conn),
	}
}
