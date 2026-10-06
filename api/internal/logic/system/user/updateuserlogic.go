// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package user

import (
	"context"
	"database/sql"

	"github.com/feihua/zero-admin-template/api/internal/common/errorx"
	"github.com/feihua/zero-admin-template/api/internal/model"
	"github.com/feihua/zero-admin-template/api/internal/svc"
	"github.com/feihua/zero-admin-template/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserReq) (resp *types.UpdateUserResp, err error) {
	resp = &types.UpdateUserResp{}
	err = l.svcCtx.UserModel.Update(l.ctx, &model.SysUser{
		Id:       req.Id,
		Mobile:   req.Mobile,
		UserName: req.UserName,
		StatusId: req.StatusId,
		Sort:     req.Sort,
		Remark:   sql.NullString{String: req.Remark, Valid: true},
	})

	if err != nil {
		return nil, errorx.NewDefaultError("更新用户异常")
	}

	resp.Code = 0
	resp.Msg = "更新成功"
	return
}
