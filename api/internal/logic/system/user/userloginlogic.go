// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package user

import (
	"context"
	"strings"
	"time"

	"github.com/feihua/zero-admin-template/api/internal/common/errorx"
	"github.com/feihua/zero-admin-template/api/internal/svc"
	"github.com/feihua/zero-admin-template/api/internal/types"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/stores/sqlc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLoginLogic {
	return &UserLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLoginLogic) UserLogin(req *types.UserLoginReq) (resp *types.UserLoginResp, err error) {
	resp = &types.UserLoginResp{}
	if len(strings.TrimSpace(req.Mobile)) == 0 || len(strings.TrimSpace(req.Password)) == 0 {
		return nil, errorx.NewDefaultError("参数错误")
	}

	userInfo, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, req.Mobile)
	switch err {
	case nil:
	case sqlc.ErrNotFound:
		logx.WithContext(l.ctx).Errorf("用户不存在 参数:%s 异常:%s", req.Mobile, err.Error())
		return nil, errorx.NewDefaultError("用户不存在")
	default:
		logx.WithContext(l.ctx).Errorf("用户登录失败 参数:%s 异常:%s", req.Mobile, err.Error())
		return nil, errorx.NewDefaultError("用户登录失败")
	}

	if userInfo.Password != req.Password {
		logx.WithContext(l.ctx).Errorf("用户密码不正确 参数:%s", req.Mobile)
		return nil, errorx.NewDefaultError("用户密码不正确")
	}

	now := time.Now().Unix()
	accessSecret := l.svcCtx.Config.Auth.AccessSecret
	accessExpire := l.svcCtx.Config.Auth.AccessExpire

	jwtToken, _ := l.getJwtToken(accessSecret, userInfo.UserName, now, accessExpire, userInfo.Id)

	resp.Code = 0
	resp.Msg = "登录成功"
	resp.Data = types.UserLoginData{
		Mobile: req.Mobile,
		Token:  jwtToken,
		UserId: 1,
	}
	return
}

func (l *UserLoginLogic) getJwtToken(secretKey, userName string, iat, seconds, userId int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["userId"] = userId
	claims["userName"] = userName
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
