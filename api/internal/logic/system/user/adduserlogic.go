// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/feihua/zero-admin-template/api/internal/common/errorx"
	"github.com/feihua/zero-admin-template/api/internal/model"
	"github.com/feihua/zero-admin-template/api/internal/svc"
	"github.com/feihua/zero-admin-template/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddUserLogic {
	return &AddUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddUserLogic) AddUser(req *types.AddUserReq) (resp *types.AddUserResp, err error) {
	mobile, _ := l.svcCtx.UserModel.FindOneByMobile(l.ctx, req.Mobile)
	if mobile != nil {
		return nil, errorx.NewDefaultError("mobile already exists")
	}

	_, err = l.svcCtx.UserModel.Insert(l.ctx, &model.SysUser{
		Mobile:   req.Mobile,
		UserName: req.UserName,
		Password: "123456",
		StatusId: req.StatusId,
		Sort:     req.Sort,
		Remark: sql.NullString{
			String: req.Remark,
			Valid:  true,
		},
		CreateTime: time.Time{},
		UpdateTime: sql.NullTime{},
	})

	if err != nil {
		return nil, errorx.NewDefaultError("添加用户异常")
	}

	resp = &types.AddUserResp{}
	resp.Code = 0
	resp.Msg = "添加用户成功"
	return
}
