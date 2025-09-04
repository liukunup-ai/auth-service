package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	Auth struct {
		AccessSecret string
		AccessExpire int64
	}

	Mysql struct {
		DataSource string
	}

	Redis struct {
		Addrs    []string
		DB       int
		Password string
	}

	Captcha struct {
		Enable bool
		Expire int64
		Length int
	}
}
