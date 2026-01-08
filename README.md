# Common Auth Service

一个简单的认证服务，基于 go-zero 框架构建。

## 🔐 功能模块

- **认证**
  - 注册/登录/登出
  - 验证码
  - 令牌刷新 (Double Token)
  - 令牌验证
- **用户管理**
  - 获取/修改用户信息
  - 修改/重置密码
- **权限管理**
  - 角色查询
  - 权限验证
- **管理接口**
  - 用户列表
  - 用户状态管理
  - 角色分配

## 🛡️ 安全特性

- 支持 **验证码** 防护
- 支持 **JWT** 认证 (AccessToken + RefreshToken 自动刷新机制)
- 支持 **密码强度验证**

## 🛠️ 技术栈

- [go-zero](https://go-zero.dev/)
- MySQL

## 🚀 常用命令

```shell
# 确保在 projects 的根目录下
cd auth-service

# 生成 HTTP 接口代码
goctl api go --api ./dsl/auth.api --dir ./ --style goZero

# 生成 MySQL 模型代码
goctl model mysql ddl --src ./model/mysql/user.sql --dir ./model/mysql

# 运行服务
go run auth.go -f etc/auth-api.yaml
```

## 📂 目录结构

```plaintext
.
├── dsl/                  # 接口定义 (DSL)
│   ├── auth.api          # Auth 服务接口定义
│   └── ...
├── etc/                  # 配置文件
│   └── auth-api.yaml     # 服务配置
├── http/                 # HTTP 请求示例
├── internal/             # 内部逻辑 (go-zero 生成)
│   ├── config/           # 配置定义
│   ├── handler/          # 路由处理
│   ├── logic/            # 业务逻辑
│   ├── middleware/       # 中间件
│   ├── svc/              # 服务上下文
│   └── types/            # 类型定义
├── model/                # 数据模型
│   └── mysql/            # MySQL 模型
├── tests/                # 测试用例
├── auth.go               # 服务入口
├── Makefile              # 便捷命令
├── go.mod
└── README.md
```

## 📦 安装与使用

1. 安装依赖: `go mod tidy`
2. 启动依赖服务 (MySQL, Redis 等)
3. 运行服务: `make run` 或 `go run auth.go -f etc/auth-api.yaml`
