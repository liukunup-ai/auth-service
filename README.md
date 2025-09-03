# Common Auth Service

## go-zero

```
go version

go install github.com/zeromicro/go-zero/tools/goctl@latest

goctl --version
```

auth-service/
├── api
│   ├── auth.api          # 你的API定义文件
│   ├── auth.go           # main函数入口
│   ├── etc
│   │   └── auth-api.yaml # 配置文件
│   └── internal/
│       ├── config/       # 配置定义
│       ├── handler/      # 路由处理器
│       ├── logic/        # 业务逻辑
│       ├── middleware/   # 中间件
│       ├── svc/          # 服务上下文
│       └── types/        # 请求/响应类型
├── go.mod
└── go.sum

# 在 auth-service 目录下执行
# 使用 goZero 风格（注意大小写，根据你goctl版本支持的模式）
goctl api go -api auth.api -dir . -style goZero

goctl api go -api ./api/auth.api -dir ./api -style goZero

# 在rpc目录下执行
goctl rpc protoc auth.proto --go_out=. --go-grpc_out=. --zrpc_out=. -style=goZero

xiaoxin-technology.goctl

🔐 主要功能模块
基础认证接口

登录/登出

注册

令牌刷新

令牌验证

用户管理接口

获取用户信息

修改个人信息

修改密码

重置密码

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
