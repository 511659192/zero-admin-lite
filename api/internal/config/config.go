// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	Mysql struct {
		Datasource string
	}

	Auth struct {
		AccessSecret string
		AccessExpire int64
	}

	CacheRedis cache.CacheConf

	Redis struct {
		Address string
		Pass    string
	}
}
