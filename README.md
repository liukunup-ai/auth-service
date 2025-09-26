# Common Auth Service

一个简单的认证服务

## 🔐 功能模块

- 验证码
- 注册
- 登录/登出
- 令牌刷新
- 令牌验证
- 修改密码
- 邮箱重置密码

基础认证接口

用户管理接口

获取用户信息

修改个人信息

安全验证接口

权限验证

角色查询

管理接口（需要管理员权限）

用户列表

用户状态管理

角色分配

## 🛡️ 安全特性

- 支持`验证码`
- 支持`JWT`认证（实现了`accessToken`+`refreshToken`双Token滚动刷新机制）
- 支持`密码强度验证`
- 支持`xx`



## 🚀 使用方式
保存为 auth.api 文件

使用 goctl 生成代码：

bash
goctl api go -api auth.api -dir . -style goZero
这个设计考虑了生产环境的需求，包括安全验证、权限管理和可扩展性。


## 技术栈

- [go-zero](https://go-zero.dev/)

## 常用命令

```shell
# 确保在项目的根目录下
cd auth-service

# HTTP
goctl api go --api ./api/dsl/auth.api --dir ./api/ --style goZero
# 运行
cd api
go run auth.go

# RPC
goctl rpc protoc ./rpc/dsl/auth.proto --go_out=./rpc --go-grpc_out=./rpc --zrpc_out=./rpc --style goZero

# MySQL
goctl model mysql ddl --src ./model/mysql/user.sql --dir ./model/mysql
```

## 目录结构

```plaintext
.
├── api                   # HTTP 服务
│   ├── dsl               # 在这里设计和定义你的 HTTP 接口
│   │   └── auth.api      #
│   ├── http/             # HTTP Client 接口测试用例
│   ├── etc               # 配置文件
│   │   └── auth-api.yaml #
│   ├── internal          # 生成的代码
│   │   ├── config/       # 配置类
│   │   ├── handler/      #
│   │   ├── logic/        # 业务逻辑
│   │   ├── middleware/   # 中间件
│   │   ├── svc/          #
│   │   └── types/        #
│   └── auth.go           # 服务入口
├── rpc                   # RPC 服务
│   ├── dsl               # 在这里设计和定义你的 RPC 接口
│   │   └── auth.proto    #
│   ├── etc               # 配置文件
│   │   └── auth.yaml     #
│   ├── internal          #
│   │   ├── config/       #
│   │   ├── logic/        #
│   │   ├── server/       #
│   │   └── svc/          #
│   ├── auth/             # *.pb.go 文件 (请勿修改)
│   ├── authClient/       # RPC 客户端
│   └── auth.go           # 服务入口
├── model                 #
│   └── mysql             #
│       └── user.sql      #
├── util                  #
├── deploy                #
├── Makefile              # 便捷命令
├── *.code-workspace      # 工作空间的配置文件
├── go.mod                #
├── go.sum                #
├── .gitignore            #
├── LICENSE               #
└── README.md             #
```





## go-zero

```
go version

go install github.com/zeromicro/go-zero/tools/goctl@latest

goctl --version
```










xiaoxin-technology.goctl

🔐 主要功能模块

- 认证（基础 + LDAP + OIDC）
  - 验证码（默认开启，支持关闭）
  - 注册
  - 登录
  - 登出
  - 刷新令牌（双Token）
  - 验证令牌（可提供端点给apisix这样的网关）
  - 修改密码
  - 忘记密码（重置密码）

- 权限（RBAC）










令牌刷新

令牌验证

用户管理接口

获取用户信息

修改个人信息

修改密码

安全验证接口

验证码获取

权限验证

角色查询

管理接口（需要管理员权限）

用户列表

用户状态管理

角色分配

🛡️ 安全特性
JWT 令牌认证

验证码保护

密码强度验证

权限层级控制

🚀 使用方式
保存为 auth.api 文件

使用 goctl 生成代码：

bash
goctl api go -api auth.api -dir . -style goZero
这个设计考虑了生产环境的需求，包括安全验证、权限管理和可扩展性。
