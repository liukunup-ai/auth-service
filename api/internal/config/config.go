package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	Mysql struct {
		DataSource string
	}

	Redis struct {
		Addrs    []string
		DB       int
		Password string
	}

	Auth struct {
		AccessSecret         string
		AccessExpiresIn      int64
		RefreshSecret        string
		RefreshExpiresIn     int64
		BlacklistCachePrefix string
	}

	Captcha struct {
		Enable      bool
		ExpiresIn   int64
		Length      int
		CachePrefix string
	}
}
