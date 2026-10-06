// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"github.com/feihua/zero-admin-template/api/internal/config"
	"github.com/feihua/zero-admin-template/api/internal/middleware"
	"github.com/feihua/zero-admin-template/api/internal/model"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config    config.Config
	CheckUrl  rest.Middleware
	UserModel model.SysUserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.Mysql.Datasource)
	return &ServiceContext{
		Config:    c,
		CheckUrl:  middleware.NewCheckUrlMiddleware().Handle,
		UserModel: model.NewSysUserModel(sqlConn, c.CacheRedis),
	}
}
