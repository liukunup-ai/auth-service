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
		AccessSecret string
		AccessExpire int64
	}

	Captcha struct {
		Enable bool
		Expire int64
		Length int
	}
}
