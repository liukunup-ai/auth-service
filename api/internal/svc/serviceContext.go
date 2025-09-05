package svc

import (
	"context"
	"fmt"
	"time"

	"auth-service/api/internal/config"
	"auth-service/api/internal/middleware"
	model "auth-service/model/mysql"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config          config.Config
	DB              sqlx.SqlConn
	Redis           redis.UniversalClient
	PasswordEncoder *PasswordEncoder
	Captcha         *base64Captcha.Captcha
	JWT             *JWT
	AuthInterceptor rest.Middleware
	UserModel       model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	var (
		err error

		db      sqlx.SqlConn
		rdb     redis.UniversalClient
		captcha *base64Captcha.Captcha
	)

	if db, err = initDatabase(c); err != nil {
		panic(fmt.Sprintf("failed to initialize database: %v", err))
	}

	if rdb, err = initRedis(c); err != nil {
		panic(fmt.Sprintf("failed to initialize redis: %v", err))
	}

	if captcha, err = initCaptcha(c, rdb); err != nil {
		panic(fmt.Sprintf("failed to initialize captcha: %v", err))
	}

	return &ServiceContext{
		Config:          c,
		DB:              db,
		Redis:           rdb,
		PasswordEncoder: &PasswordEncoder{},
		Captcha:         captcha,
		JWT:             NewJWT(c, rdb),
		AuthInterceptor: middleware.NewAuthInterceptorMiddleware().Handle,
		UserModel:       model.NewUserModel(db),
	}
}

func initDatabase(c config.Config) (sqlx.SqlConn, error) {
	// 初始化 MySQL 连接
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	// 测试连接
	var result int64
	if err := conn.QueryRow(&result, "SELECT 1"); err != nil || result != 1 {
		return nil, fmt.Errorf("failed to connect to mysql: %v", err)
	}
	return conn, nil
}

func initRedis(c config.Config) (redis.UniversalClient, error) {
	// 初始化 Redis 客户端
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    c.Redis.Addrs,
		DB:       c.Redis.DB,
		Password: c.Redis.Password,
	})
	// 测试连接
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}
	return rdb, nil
}

func initCaptcha(c config.Config, rdb redis.UniversalClient) (*base64Captcha.Captcha, error) {
	// 创建驱动 (使用数字验证码)
	driver := base64Captcha.NewDriverDigit(80, 240, c.Captcha.Length, 0.7, 80)
	if driver == nil {
		return nil, fmt.Errorf("failed to create captcha driver")
	}
	// 使用 Redis 作为存储
	store := NewRedisStore(context.Background(), rdb, c.Captcha.CachePrefix, time.Duration(c.Captcha.Expire)*time.Second)
	if store == nil {
		return nil, fmt.Errorf("failed to create captcha store")
	}
	return base64Captcha.NewCaptcha(driver, store), nil
}
