// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package user

import (
	"context"
	"time"

	"github.com/feihua/zero-admin-template/api/internal/common/errorx"
	"github.com/feihua/zero-admin-template/api/internal/svc"
	"github.com/feihua/zero-admin-template/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryUserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserListLogic {
	return &QueryUserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryUserListLogic) QueryUserList(req *types.QueryUserListReq) (resp *types.QueryUserListResp, err error) {
	resp = &types.QueryUserListResp{}

	userList, count, err := l.svcCtx.UserModel.FindAll(l.ctx, req.Current, req.PageSize, req.Mobile, req.StatusId)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("查询用户列表异常: %s", err.Error())
		return nil, errorx.NewDefaultError("查询用户列表异常")
	}

	var list []types.UserListData
	for _, user := range *userList {
		list = append(list, types.UserListData{
			Id:         user.Id,
			Mobile:     user.Mobile,
			UserName:   user.UserName,
			StatusId:   user.StatusId,
			Sort:       user.Sort,
			Remark:     user.Remark.String,
			CreateTime: user.CreateTime.Format(time.DateTime),
			UpdateTime: user.UpdateTime.Time.Format(time.DateTime),
		})
	}

	resp.Code = 0
	resp.Msg = "查询成功"
	resp.Data = list
	resp.Success = true
	resp.Total = count
	return
}
