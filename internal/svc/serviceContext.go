package svc

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/middleware"
	model "auth-service/model/mysql"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"github.com/sony/sonyflake"
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
	Sonyflake       *sonyflake.Sonyflake
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
		Sonyflake:       initSonyflake(),
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
	store := NewRedisStore(context.Background(), rdb, c.Captcha.CachePrefix, time.Duration(c.Captcha.ExpiresIn)*time.Second)
	if store == nil {
		return nil, fmt.Errorf("failed to create captcha store")
	}
	return base64Captcha.NewCaptcha(driver, store), nil
}

func initSonyflake() *sonyflake.Sonyflake {
	settings := sonyflake.Settings{
		StartTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		MachineID: getMachineID,
		CheckMachineID: func(id uint16) bool {
			return id < 1024
		},
	}
	return sonyflake.NewSonyflake(settings)
}

func getMachineID() (uint16, error) {
	// 1. 从环境变量获取
	if machineID := os.Getenv("SNOWFLAKE_MACHINE_ID"); machineID != "" {
		if id, err := strconv.ParseUint(machineID, 10, 16); err == nil {
			if id < 1024 {
				fmt.Printf("从环境变量获取到 MachineID: %d\n", id)
				return uint16(id), nil
			}
			fmt.Printf("环境变量 MachineID 超出范围: %d, 使用备用方案\n", id)
		} else {
			fmt.Printf("环境变量 MachineID 格式错误: %s, 使用备用方案\n", machineID)
		}
	}

	// 2. 基于 IP 地址生成
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return 0, fmt.Errorf("获取网络接口失败: %v", err)
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip := ipNet.IP.To4(); ip != nil {
				// 使用 IP 地址的最后两个字节生成 MachineID
				// 例如: 192.168.1.100 -> 1*256 + 100 = 356
				machineID := (uint16(ip[2])<<8 + uint16(ip[3])) % 1024
				fmt.Printf("基于 IP 地址生成 MachineID: %d\n", machineID)
				return machineID, nil
			}
		}
	}

	// 3. 回退到随机数 (使用时间种子)
	rand.New(rand.NewSource(time.Now().UnixNano()))
	machineID := uint16(rand.Intn(1024))
	fmt.Printf("使用随机数作为 MachineID: %d\n", machineID)
	return machineID, nil
}
