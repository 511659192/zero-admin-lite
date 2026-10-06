# 从零到一:用 go-zero 搭建一个带 JWT 鉴权的用户管理服务(zero-admin-lite 实录)

> 本文基于 zeor-admin-lite 仓库的真实 git 历史还原(从第一个提交到接口联调通过,同一天完成)。
> 每一步都是实际走过的路,包括踩过的坑——新手上手时大概率会遇到一模一样的报错。

## 项目简介

**zero-admin-lite** 是一个最小化的后台管理系统 API 服务:

- 技术栈:go-zero v1.6.2 + goctl 1.10.2 + MySQL + Redis
- 功能:用户登录(JWT 鉴权)+ 用户的增删改查(RBAC 权限体系的第一块基石)
- 开发模式:**契约先行(Contract-First)** —— 先写 `.api` 接口描述文件,用 goctl 一键生成代码,再往里填业务逻辑

最终实现并调通的 5 个接口:

| 接口 | 路径 | 鉴权 |
|---|---|---|
| 用户登录 | `POST /api/system/user/login` | 公开 |
| 查询用户列表 | `POST /api/system/user/queryUserList` | JWT |
| 新增用户 | `POST /api/system/user/addUser` | JWT |
| 更新用户 | `POST /api/system/user/updateUser` | JWT |
| 删除用户 | `POST /api/system/user/deleteUser` | JWT |

---

## 第 0 步:环境准备

| 工具 | 说明 | 安装 |
|---|---|---|
| Go 1.21+ | 运行时 | 官网下载 |
| goctl | go-zero 的代码生成器 | `go install github.com/zeromicro/go-zero/tools/goctl@latest` |
| MySQL 8.x | 业务库(库名 `hertzdb`) | 本地安装即可 |
| Redis | model 缓存层 | 本地安装即可 |

验证:

```bash
go version   # go1.21+
goctl --version  # 1.10.2
redis-cli ping   # PONG
```

---

## 第 1 步:定义 API 契约(`api/doc/*.api`)

go-zero 用 `.api` 文件描述"服务长什么样":路由、请求/响应结构、中间件、JWT。这是整个项目的**源头**。

一个精简后的 `api/doc/system/user.api` 长这样(节选):

```api
syntax = "v1"

info (
	title:   "用户信息"
	desc:    "简单实现用户管理"
	version: "v1.0"
)

// ---- 请求/响应结构体 ----
type AddUserReq {
	Mobile   string `json:"mobile"`
	UserName string `json:"userName"`
	StatusId int64  `json:"statusId"`
	Sort     int64  `json:"sort"`
	Remark   string `json:"remark"`
}

type QueryUserListReq {
	Current  int64  `json:"current,default=1"`
	PageSize int64  `json:"pageSize,default=20"`
	Mobile   string `json:"mobile,optional"`
	StatusId int64  `json:"statusId,default=2"`
}

// ---- 受 JWT 保护的路由组 ----
@server (
	jwt:        Auth              // 引用 config 里的 Auth,自动挂 JWT 中间件
	middleware: CheckUrl          // 自定义中间件
	group:      system/user       // 生成代码的子目录
	prefix:     /api/system/user  // 路由前缀
)
service admin {
	@handler addUser        // 生成 handler/logic 的名字
	post /addUser (AddUserReq) returns (AddUserResp)
	post /deleteUser (DeleteUserReq) returns (DeleteUserResp)
	post /updateUser (UpdateUserReq) returns (UpdateUserResp)
	post /queryUserList (QueryUserListReq) returns (QueryUserListResp)
}

// ---- 公开的登录路由(不挂 jwt) ----
@server (
	group:  system/user
	prefix: /api/system
)
service admin {
	@handler userLogin
	post /user/login (UserLoginReq) returns (UserLoginResp)
}
```

入口文件 `api/doc/admin.api` 用 import 组合多个模块:

```api
import "system/user.api"

info (
	title:   "zero-admin-lite"
	desc:    "简单实现用户管理"
	version: "v1.0"
)
```

> 💡 **实际做法(可借鉴)**:本项目最初把参考项目 zero-admin-template 的**全部模块**(user/role/menu/登录日志/操作日志,约 700 行 api)一次性抄过来生成了 66 个文件,提交一;然后**删除** role/menu/log 等模块只保留 user,提交二;再重新生成,提交三。这就是"先全量、再瘦身"的路线——好处是能看到完整模块的规范写法,坏处是绕了一步。**新手建议直接从单模块 user.api 起步**,以后加模块就是加一个 `.api` 文件。

> 💡 `.api` 里 `json` tag 支持 `optional`(可不传)、`default=xx`(缺省值),goctl 生成代码时会自动做参数校验和默认值填充,登录接口的 `Mobile string \`json:"account"\`` 就是故意把 JSON 字段名映射成 account。

---

## 第 2 步:一行命令生成整个服务骨架

```bash
cd api
goctl api go -api doc/admin.api -dir .          # -dir 指定生成根目录
go mod tidy                                      # 补齐依赖
```

生成后的目录结构(每层职责要记住):

```
api/
├── admin.go                     # main 入口:加载配置 → 建路由 → 启动
├── etc/
│   └── admin.yaml               # 配置文件(下一步)
├── doc/                         # .api 契约文件(源码的一部分,要提交)
└── internal/
    ├── config/config.go         # 配置结构体(与 yaml 一一对应)
    ├── handler/                 # 入参解析/响应写出,不许写业务
    ├── logic/                   # ★ 业务逻辑写在这里(唯一要你填的层)
    ├── types/types.go           # 由 .api 生成的请求/响应结构体
    ├── svc/servicecontext.go    # 依赖容器:config、model、redis 都挂这里
    ├── middleware/              # 自定义中间件
    └── common/errorx/           # 统一业务错误(后加的)
```

生成完直接就能跑(虽然所有 logic 都是空壳):

```bash
go run admin.go -f etc/admin.yaml
```

---

## 第 3 步:补配置文件 —— ⚠️ 第一个坑

第一次启动大概率会看到:

```
error: config file etc/admin.yaml, "Auth" is not set
```

**原因**:`.api` 里声明了 `jwt: Auth`,goctl 生成的 `config.go` 就包含 Auth 必填字段,但脚手架默认的 yaml 里没有。**配置结构体与 yaml 是严格对应的**,补上即可:

```yaml
Name: admin
Host: 0.0.0.0
Port: 7778

Auth:
  AccessSecret: "改成你的随机长字符串"   # JWT 签名密钥
  AccessExpire: 86400                    # token 有效期(秒)
```

后续接入数据库和缓存后,最终版配置:

```yaml
Mysql:
  Datasource: root:root@tcp(127.0.0.1:3306)/hertzdb?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai

Auth:
  AccessSecret: "你的密钥"
  AccessExpire: 86400

CacheRedis:                # model 缓存用(--cache 生成时必填)
  - Host: 127.0.0.1:6379
    Pass: root
    Type: node

Redis:
  Address: 127.0.0.1:6379  # 业务自定义 Redis(可选)
  Pass: root
```

> ⚠️ yaml 键与 Go 字段名大小写不敏感但拼写必须一致;段名就是 config 结构体里的字段名。

---

## 第 4 步:建表 + 生成数据层(model)

### 4.1 写 DDL(`api/doc/init.sql`)

```sql
CREATE TABLE `sys_user` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
    `mobile` char(11) NOT NULL DEFAULT '' COMMENT '手机',
    `user_name` varchar(50) NOT NULL COMMENT '姓名',
    `password` varchar(64) NOT NULL COMMENT '密码',
    `status_id` tinyint NOT NULL DEFAULT '1' COMMENT '状态(1:正常,0:禁用)',
    `sort` int NOT NULL DEFAULT '1' COMMENT '排序',
    `remark` varchar(255) DEFAULT NULL COMMENT '备注',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `AK_phone` (`mobile`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户信息';
```

### 4.2 生成 model

```bash
goctl model mysql ddl --src api/doc/init.sql --dir api/internal/model --cache --database hertzdb
```

这条命令是整个项目**参数最多、坑最多**的一步,逐个解释:

| 参数 | 作用 | 不加会怎样 |
|---|---|---|
| `--src` | DDL 文件路径,一个文件里放多个 CREATE TABLE 完全可以 | - |
| `--dir` | 输出目录 | - |
| `--cache` | 生成带 Redis 缓存的 model(查询走缓存,写操作自动删缓存) | 不加则是纯直连版 |
| `--database hertzdb` | 把库名写进缓存 key,如 `cache:hertzdb:sysUser:id:` | 不加 key 是 `cache:sysUser:id:`,多服务共用 Redis 时可能撞 key |

生成产物:

```
model/
├── sysusermodel.go       # 自定义层(git 历史里的"Safe to edit"),手写业务方法
├── sysusermodel_gen.go   # 生成层(DO NOT EDIT),CRUD + 缓存逻辑
└── vars.go               # ErrNotFound 等公共定义
```

### 4.3 这一过程中的坑(全部真实踩过)

**坑 ① DDL 语法错误导致解析失败**

```
Error: init.sql line 13:103 mismatched input ')'
```

从文档复制 SQL 时带入了多余括号:`CURRENT_TIMESTAMP) COMMENT`、表定义结尾多一个 `)`。goctl 用严格的 DDL 解析器,一个字符都不能错。**对策:先在 MySQL 里执行一遍确认 SQL 无误,再喂给 goctl。**

**坑 ② 重新生成不会覆盖自定义文件**

```
sysusermodel.go already exists, ignored.
```

goctl 的设计:`_gen.go` 每次重写,`xxxmodel.go`(自定义层)已存在则跳过。所以**手写的方法要放自定义层**,不会被冲掉。

**坑 ③ 生成后编译缺驱动**

```
missing go.sum entry for module providing package github.com/go-sql-driver/mysql
```

`go mod tidy` 一下就好。

---

## 第 5 步:装配 ServiceContext(依赖容器)

`svc/servicecontext.go` 是给 logic 层"发工具"的地方:

```go
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
		UserModel: model.NewSysUserModel(sqlConn, c.CacheRedis),   // 缓存版 model 必须传 CacheRedis
	}
}
```

> ⚠️ `--cache` 生成的 model 构造函数是 `NewSysUserModel(conn, c cache.CacheConf, opts ...)`,和直连版签名不同。生成后 **ServiceContext 要同步改**,否则编译报 `not enough arguments`。

---

## 第 6 步:实现业务逻辑(logic 层)

### 6.1 登录 + JWT 签发(`userloginlogic.go`)

```go
func (l *UserLoginLogic) UserLogin(req *types.UserLoginReq) (resp *types.UserLoginResp, err error) {
	// 1. 参数校验
	// 2. FindOneByMobile 查用户
	// 3. 校验密码(示例为明文,生产必须 bcrypt)
	// 4. 签发 JWT
	now := time.Now().Unix()
	jwtToken, _ := l.getJwtToken(l.svcCtx.Config.Auth.AccessSecret, userInfo.UserName, now, l.svcCtx.Config.Auth.AccessExpire, userInfo.Id)
	...
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
```

`.api` 里的 `jwt: Auth` + 这里的 `AccessSecret` 是同一把钥匙:客户端之后带 `Authorization: Bearer <token>`,go-zero 自动验签,并把 `userId`/`userName` 塞进请求 context 供后续 logic 取用。

### 6.2 统一业务错误(`common/errorx/baseerror.go`)

```go
const defaultCode = 1001

type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewDefaultError(msg string) error { return NewCodeError(defaultCode, msg) }
```

logic 里 `return nil, errorx.NewDefaultError("用户不存在")`,handler 自动把它变成 `{"code":1001,"msg":"用户不存在"}`。

### 6.3 分页查询(⚠️ 最大的坑在这里)

在**自定义层** `sysusermodel.go` 加业务方法(生成层不许改):

```go
func (m *defaultSysUserModel) FindAll(ctx context.Context, current, pageSize int64, mobile string, statusId int64) (*[]SysUser, int64, error) {
	where := "1=1"
	if len(mobile) > 0 {
		where += fmt.Sprintf(" AND (user_name like '%%%s%%' OR mobile like '%%%s%%')", mobile, mobile)
	}
	if statusId != 2 {
		where += fmt.Sprintf(" AND status_id = %d", statusId)
	}
	where += " ORDER BY create_time DESC"

	var resp []SysUser
	query := fmt.Sprintf("select %s from %s where %s limit ?, ?", sysUserRows, m.table, where)
	// ★ 多行查询必须用 QueryRows(复数)!
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, (current-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var count int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &count, countQuery); err != nil {   // 单行用 QueryRow
		return nil, 0, err
	}
	return &resp, count, nil
}
```

> ⚠️ **本项目真实踩坑实录**:列表查询误用了单行的 `QueryRowNoCacheCtx`,把多行结果扫进 `[]SysUser`,报错 `unsupported unmarshal type`,而且 logic 层把原始错误吞掉只返回了"查询用户列表异常",排查绕了远路。两条教训:
> 1. **`QueryRow` = 单行,`QueryRows` = 多行**,和返回目标是 `*T` 还是 `*[]T` 必须配对;
> 2. **logic 里捕获 err 后一定要先 `logx.Errorf` 打日志再返回友好错误**,否则线上根本不知道发生了什么。

### 6.4 CRUD 的其余三个

`addUser`(先查重 mobile 再 Insert)、`updateUser`、`deleteUser`(支持批量,循环调用 Delete),套路一致:调 model → 判断 err → 包装 CodeError。注意 goctl 默认忽略 `create_time`/`update_time` 列,交给数据库维护。

---

## 第 7 步:启动与联调

```bash
go build -o zero-admin.exe . && ./zero-admin.exe -f etc/admin.yaml
```

用 curl 全链路验证(先登录拿 token,再带 token 调受保护接口):

```bash
# 1. 登录
curl -X POST http://127.0.0.1:7778/api/system/user/login \
  -H "Content-Type: application/json" \
  -d '{"account":"18613030352","password":"123456"}'
# → {"code":0,"msg":"登录成功","data":{"token":"eyJhbG..."}} 

# 2. 查询列表(替换 <token>)
curl -X POST http://127.0.0.1:7778/api/system/user/queryUserList \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"current":1,"pageSize":10}'
# → {"code":0,"msg":"查询成功","data":[...],"success":true,"total":2}
```

### 联调期补充踩坑

| 现象 | 原因 | 处理 |
|---|---|---|
| `bind: Only one usage of each socket address` | 端口被上一个进程占用 | 改 Port 或 `taskkill /F /IM xxx.exe` |
| JWT 里 userName 是乱码 | Redis 里留着旧进程写入的脏缓存 | `redis-cli del` 对应缓存 key,重新从库加载 |
| 时间显示 `0001-01-01` | `sql.NullTime` 零值被直接 Format | 输出前判断 `.Valid` |
| 请求报 `field "account" is not set` | goctl 按必填校验,字段名以 types.go 的 json tag 为准 | 对齐契约字段名 |

---

## 踩坑清单汇总(速查表)

| # | 报错/现象 | 根因 | 一句话解法 |
|---|---|---|---|
| 1 | `"Auth" is not set` | yaml 缺配置段 | config 结构体与 yaml 严格对应,补齐 |
| 2 | `mismatched input ')'` | DDL 文件语法错误 | 先在 MySQL 验证 SQL 再喂 goctl |
| 3 | `already exists, ignored` | 重复生成 | 正常现象,自定义层不会被覆盖 |
| 4 | `missing go.sum entry` | 依赖未入账 | `go mod tidy` |
| 5 | `not enough arguments` | `--cache` 版 model 构造签名变了 | ServiceContext 传 `c.CacheRedis` |
| 6 | `unsupported unmarshal type` | `QueryRow` 扫多行结果 | 多行用 `QueryRows` |
| 7 | 缓存 key 没带库名 | `--src` 模式无库上下文 | 加 `--database <库名>` |
| 8 | 端口占用 | 旧进程未停 | 杀进程或换端口 |
| 9 | JWT 中文乱码 | Redis 脏缓存 | 删 key 重建缓存 |

## 常用命令速查

```bash
# 由 .api 生成 API 服务骨架
goctl api go -api api/doc/admin.api -dir api

# 由 DDL 生成 model(推荐带缓存版)
goctl model mysql ddl --src api/doc/init.sql --dir api/internal/model --cache --database hertzdb

# 看某表已有哪些路由(启动时)
./zero-admin.exe -f etc/admin.yaml   # 启动即打印全部路由
```

## 总结:goctl 的"契约先行"开发循环

```
改 .api(契约) ──→ goctl api go 重新生成 ──→ 生成层自动更新、自定义层保留
改 DDL(表结构) ──→ goctl model 重新生成 ──→ _gen.go 自动更新、手写方法保留
```

- **只往 `logic` 和自定义 `model` 里写代码**,其余全是生成物
- 契约文件(`.api` + `init.sql`)就是文档,入库 git
- 新手最容易犯的两个错:配置/生成物不同步(编译错)、QueryRow/QueryRows 用混(运行时错)——记住这两点,能少走 90% 弯路
