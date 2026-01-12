# APISIX JWT + OIDC 集成指南

本文档介绍如何将 `auth-service` 与 Apache APISIX 的 JWT 插件和 OIDC 插件进行集成，实现统一的认证网关。

## 目录

- [架构概述](#架构概述)
- [方案一：APISIX JWT 插件 + auth-service](#方案一apisix-jwt-插件--auth-service)
- [方案二：APISIX OIDC 插件](#方案二apisix-oidc-插件)
- [方案三：混合模式](#方案三混合模式)
- [配置示例](#配置示例)

---

## 架构概述

```
                                    ┌─────────────────┐
                                    │   OIDC Provider │
                                    │ (Keycloak/Okta) │
                                    └────────┬────────┘
                                             │
    ┌──────────┐     ┌──────────────┐     ┌──┴───────────┐
    │  Client  │────▶│    APISIX    │────▶│ auth-service │
    │  (Web)   │◀────│   Gateway    │◀────│   (Backend)  │
    └──────────┘     └──────────────┘     └──────────────┘
                            │
                            ▼
                     ┌──────────────┐
                     │   Backend    │
                     │   Services   │
                     └──────────────┘
```

---

## 方案一：APISIX JWT 插件 + auth-service

### 场景说明

使用 `auth-service` 作为 JWT 签发服务，APISIX 负责验证 JWT Token。

### 1. 配置 auth-service

确保 `auth-service` 的 JWT 配置：

```yaml
# etc/auth-api.yaml
Auth:
  AccessSecret: "your-jwt-secret-key-must-be-at-least-32-chars"
  AccessExpiresIn: 86400
  RefreshSecret: "your-refresh-secret-key-must-be-at-least-32-chars"
  RefreshExpiresIn: 604800
```

### 2. 配置 APISIX jwt-auth 插件

#### 2.1 创建 Consumer

```bash
curl -X PUT http://127.0.0.1:9180/apisix/admin/consumers \
  -H 'X-API-KEY: your-admin-key' \
  -d '{
    "username": "auth-service-consumer",
    "plugins": {
        "jwt-auth": {
            "key": "auth-service",
            "secret": "your-jwt-secret-key-must-be-at-least-32-chars",
            "algorithm": "HS256"
        }
    }
}'
```

#### 2.2 创建路由并启用 jwt-auth 插件

```bash
# 需要认证的路由
curl -X PUT http://127.0.0.1:9180/apisix/admin/routes/protected \
  -H 'X-API-KEY: your-admin-key' \
  -d '{
    "uri": "/api/v1/me",
    "upstream": {
        "type": "roundrobin",
        "nodes": {
            "auth-service:7001": 1
        }
    },
    "plugins": {
        "jwt-auth": {}
    }
}'

# 不需要认证的路由 (登录/注册等)
curl -X PUT http://127.0.0.1:9180/apisix/admin/routes/public \
  -H 'X-API-KEY: your-admin-key' \
  -d '{
    "uris": ["/api/v1/login", "/api/v1/register", "/api/v1/captcha", "/api/v1/sso/*"],
    "upstream": {
        "type": "roundrobin",
        "nodes": {
            "auth-service:7001": 1
        }
    }
}'
```

### 3. APISIX 配置文件方式

创建 `apisix/apisix.yaml`:

```yaml
routes:
  # 公开路由 - 无需认证
  - id: auth-public
    uri: /api/v1/login
    methods: ["POST"]
    upstream_id: auth-service
    
  - id: auth-register
    uri: /api/v1/register
    methods: ["POST"]
    upstream_id: auth-service
    
  - id: auth-captcha
    uri: /api/v1/captcha
    methods: ["GET"]
    upstream_id: auth-service

  - id: auth-sso
    uri: /api/v1/sso/*
    upstream_id: auth-service

  # 受保护路由 - 需要 JWT 认证
  - id: auth-protected
    uri: /api/v1/*
    upstream_id: auth-service
    plugins:
      jwt-auth:
        header: Authorization
        query: token
        cookie: token

upstreams:
  - id: auth-service
    type: roundrobin
    nodes:
      "auth-service:7001": 1

consumers:
  - username: auth-service-user
    plugins:
      jwt-auth:
        key: auth-service
        secret: "your-jwt-secret-key-must-be-at-least-32-chars"
        algorithm: HS256
```

---

## 方案二：APISIX OIDC 插件

### 场景说明

使用 APISIX 的 `openid-connect` 插件直接对接 OIDC Provider（如 Keycloak、Okta），由 APISIX 处理整个 OIDC 流程。

### 1. 配置 OIDC Provider

以 Keycloak 为例，创建 Client：

1. 登录 Keycloak Admin Console
2. 创建 Realm（如 `my-realm`）
3. 创建 Client：
   - Client ID: `apisix-gateway`
   - Client Protocol: `openid-connect`
   - Access Type: `confidential`
   - Valid Redirect URIs: `http://your-domain.com/*`
4. 获取 Client Secret

### 2. 配置 APISIX openid-connect 插件

```bash
curl -X PUT http://127.0.0.1:9180/apisix/admin/routes/oidc-protected \
  -H 'X-API-KEY: your-admin-key' \
  -d '{
    "uri": "/api/*",
    "upstream": {
        "type": "roundrobin",
        "nodes": {
            "backend-service:8080": 1
        }
    },
    "plugins": {
        "openid-connect": {
            "client_id": "apisix-gateway",
            "client_secret": "your-client-secret",
            "discovery": "https://keycloak.example.com/realms/my-realm/.well-known/openid-configuration",
            "scope": "openid profile email",
            "bearer_only": false,
            "realm": "my-realm",
            "introspection_endpoint_auth_method": "client_secret_post",
            "redirect_uri": "http://your-domain.com/callback",
            "logout_path": "/logout",
            "post_logout_redirect_uri": "http://your-domain.com/",
            "session": {
                "secret": "your-session-secret-at-least-32-chars"
            },
            "set_userinfo_header": true,
            "set_id_token_header": true,
            "set_access_token_header": true
        }
    }
}'
```

### 3. APISIX 配置文件方式

```yaml
routes:
  - id: oidc-protected-route
    uri: /api/*
    upstream_id: backend-service
    plugins:
      openid-connect:
        client_id: "apisix-gateway"
        client_secret: "${OIDC_CLIENT_SECRET}"
        discovery: "https://keycloak.example.com/realms/my-realm/.well-known/openid-configuration"
        scope: "openid profile email"
        bearer_only: false
        realm: "my-realm"
        redirect_uri: "http://your-domain.com/callback"
        logout_path: "/logout"
        post_logout_redirect_uri: "http://your-domain.com/"
        session:
          secret: "${SESSION_SECRET}"
        # 将用户信息传递给后端
        set_userinfo_header: true
        set_id_token_header: true
        set_access_token_header: true

upstreams:
  - id: backend-service
    type: roundrobin
    nodes:
      "backend:8080": 1
```

---

## 方案三：混合模式

### 场景说明

结合两种方式：
- APISIX OIDC 插件处理第三方登录（Google、Azure AD 等）
- auth-service 处理本地账号登录、LDAP 登录
- APISIX 统一验证 JWT

### 架构图

```
                    ┌─────────────────────────────────────┐
                    │              APISIX                  │
                    │  ┌─────────┐     ┌──────────────┐   │
    ┌────────┐      │  │  OIDC   │     │   jwt-auth   │   │
    │ Client │─────▶│  │ Plugin  │     │    Plugin    │   │
    └────────┘      │  └────┬────┘     └──────┬───────┘   │
                    └───────┼─────────────────┼───────────┘
                            │                 │
              ┌─────────────┘                 └─────────────┐
              ▼                                             ▼
    ┌──────────────────┐                         ┌──────────────────┐
    │  OIDC Provider   │                         │   auth-service   │
    │ (Keycloak/Okta)  │                         │   (JWT Issuer)   │
    └──────────────────┘                         └──────────────────┘
```

### 1. 配置 auth-service 支持 OIDC

```yaml
# etc/auth-api.yaml
SSO:
  DefaultProvider: local
  OIDC:
    Enabled: true
    ProviderURL: "https://keycloak.example.com/realms/my-realm"
    ClientID: "auth-service"
    ClientSecret: "your-client-secret"
    RedirectURL: "http://localhost:7001/api/v1/sso/oidc/callback"
    Scopes:
      - openid
      - profile
      - email
```

### 2. APISIX 路由配置

```yaml
routes:
  # auth-service 公开端点
  - id: auth-public
    uris:
      - /api/v1/login
      - /api/v1/register
      - /api/v1/captcha
      - /api/v1/refresh
      - /api/v1/sso/providers
      - /api/v1/sso/oidc/login
      - /api/v1/sso/oidc/callback
      - /api/v1/sso/ldap/login
    upstream_id: auth-service

  # auth-service 受保护端点
  - id: auth-protected
    uris:
      - /api/v1/me
      - /api/v1/password/change
      - /api/v1/logout
    upstream_id: auth-service
    plugins:
      jwt-auth: {}

  # 业务服务受保护端点
  - id: business-api
    uri: /api/business/*
    upstream_id: business-service
    plugins:
      jwt-auth: {}
      # 可选：添加用户信息到请求头
      proxy-rewrite:
        headers:
          set:
            X-User-ID: "$jwt_user_id"

consumers:
  - username: auth-service-jwt
    plugins:
      jwt-auth:
        key: auth-service
        secret: "your-jwt-secret-key-must-be-at-least-32-chars"
        algorithm: HS256

upstreams:
  - id: auth-service
    type: roundrobin
    nodes:
      "auth-service:7001": 1
  - id: business-service
    type: roundrobin
    nodes:
      "business:8080": 1
```

---

## 配置示例

### Docker Compose 完整示例

```yaml
# docker-compose.yaml
version: '3.8'

services:
  # APISIX Gateway
  apisix:
    image: apache/apisix:3.8.0-debian
    ports:
      - "9080:9080"   # HTTP
      - "9443:9443"   # HTTPS
      - "9180:9180"   # Admin API
    volumes:
      - ./apisix/config.yaml:/usr/local/apisix/conf/config.yaml
      - ./apisix/apisix.yaml:/usr/local/apisix/conf/apisix.yaml
    depends_on:
      - etcd
    environment:
      - APISIX_STAND_ALONE=true

  # etcd (APISIX 配置存储)
  etcd:
    image: bitnami/etcd:3.5
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:2379

  # auth-service
  auth-service:
    build: .
    ports:
      - "7001:7001"
    volumes:
      - ./etc/auth-api.yaml:/app/etc/auth-api.yaml
    depends_on:
      - mysql
      - redis

  # Keycloak (可选 - OIDC Provider)
  keycloak:
    image: quay.io/keycloak/keycloak:23.0
    ports:
      - "8080:8080"
    environment:
      - KEYCLOAK_ADMIN=admin
      - KEYCLOAK_ADMIN_PASSWORD=admin
    command: start-dev

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: auth
    ports:
      - "3306:3306"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --requirepass 123456
```

### APISIX 配置文件

```yaml
# apisix/config.yaml
apisix:
  node_listen: 9080
  enable_admin: true
  admin_key:
    - name: admin
      key: your-admin-api-key
      role: admin

deployment:
  role: traditional
  role_traditional:
    config_provider: yaml

plugin_attr:
  prometheus:
    export_addr:
      ip: "0.0.0.0"
      port: 9091
```

### 完整路由配置

```yaml
# apisix/apisix.yaml
routes:
  # ==================== 公开路由 ====================
  
  # 健康检查
  - id: healthz
    uri: /healthz
    upstream_id: auth-service

  # 登录相关
  - id: auth-login
    uri: /api/v1/login
    methods: ["POST"]
    upstream_id: auth-service

  - id: auth-register
    uri: /api/v1/register
    methods: ["POST"]
    upstream_id: auth-service

  - id: auth-captcha
    uri: /api/v1/captcha
    methods: ["GET"]
    upstream_id: auth-service

  - id: auth-refresh
    uri: /api/v1/refresh
    methods: ["POST"]
    upstream_id: auth-service

  - id: auth-password-forgot
    uri: /api/v1/password/forgot
    methods: ["POST"]
    upstream_id: auth-service

  - id: auth-password-reset
    uri: /api/v1/password/reset
    methods: ["POST"]
    upstream_id: auth-service

  # SSO 公开端点
  - id: sso-providers
    uri: /api/v1/sso/providers
    methods: ["GET"]
    upstream_id: auth-service

  - id: sso-oidc-login
    uri: /api/v1/sso/oidc/login
    methods: ["GET"]
    upstream_id: auth-service

  - id: sso-oidc-callback
    uri: /api/v1/sso/oidc/callback
    methods: ["GET"]
    upstream_id: auth-service

  - id: sso-ldap-login
    uri: /api/v1/sso/ldap/login
    methods: ["POST"]
    upstream_id: auth-service

  # ==================== 受保护路由 ====================
  
  - id: auth-me
    uri: /api/v1/me
    methods: ["GET"]
    upstream_id: auth-service
    plugins:
      jwt-auth: {}

  - id: auth-password-change
    uri: /api/v1/password/change
    methods: ["PUT"]
    upstream_id: auth-service
    plugins:
      jwt-auth: {}

  - id: auth-logout
    uri: /api/v1/logout
    methods: ["POST"]
    upstream_id: auth-service
    plugins:
      jwt-auth: {}

#END

upstreams:
  - id: auth-service
    type: roundrobin
    nodes:
      "auth-service:7001": 1

consumers:
  - username: auth-service-consumer
    plugins:
      jwt-auth:
        key: auth-service
        secret: "zvcbozvafjxcbdxh911bq101cblrgqbt"
        algorithm: HS256

#END
```

---

## 测试验证

### 1. 获取 JWT Token

```bash
# 本地登录获取 Token
curl -X POST http://localhost:9080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'

# 响应
{
  "userId": "xxx",
  "username": "testuser",
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 2. 访问受保护接口

```bash
# 使用 Token 访问
curl http://localhost:9080/api/v1/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."

# 成功响应
{
  "userId": "xxx",
  "username": "testuser",
  "email": "test@example.com"
}
```

### 3. SSO 登录测试

```bash
# 获取 SSO 提供者列表
curl http://localhost:9080/api/v1/sso/providers

# 响应
{
  "defaultProvider": "local",
  "providers": [
    {"id": "local", "name": "本地账号", "enabled": true},
    {"id": "oidc", "name": "OpenID Connect", "enabled": true},
    {"id": "ldap", "name": "LDAP", "enabled": true}
  ]
}

# 发起 OIDC 登录
curl http://localhost:9080/api/v1/sso/oidc/login

# 响应（重定向 URL）
{
  "authorizationUrl": "https://keycloak.example.com/realms/my-realm/protocol/openid-connect/auth?...",
  "state": "xxx"
}
```

---

## 常见问题

### Q1: JWT Secret 如何保持一致？

确保 `auth-service` 的 `Auth.AccessSecret` 与 APISIX Consumer 的 `jwt-auth.secret` 完全一致。

### Q2: 如何自定义 JWT Claims？

修改 `auth-service` 的 JWT 生成逻辑，在 `internal/svc/jwt.go` 中添加自定义 Claims：

```go
type CustomClaims struct {
    UserID   uint64   `json:"user_id"`
    Username string   `json:"username"`
    Roles    []string `json:"roles,omitempty"`  // 自定义字段
    jwt.StandardClaims
}
```

### Q3: APISIX 如何提取 JWT 中的用户信息？

使用 `consumer-restriction` 或 `serverless` 插件：

```yaml
plugins:
  jwt-auth: {}
  serverless-pre-function:
    phase: rewrite
    functions:
      - "return function(conf, ctx) 
           local jwt = require('resty.jwt')
           local token = ngx.var.http_authorization:sub(8)
           local jwt_obj = jwt:load_jwt(token)
           ngx.req.set_header('X-User-ID', jwt_obj.payload.user_id)
         end"
```

### Q4: 如何处理 Token 刷新？

客户端在 Access Token 过期前调用 `/api/v1/refresh` 接口刷新：

```bash
curl -X POST http://localhost:9080/api/v1/refresh \
  -H "Content-Type: application/json" \
  -d '{"refreshToken": "eyJhbGciOiJIUzI1NiIs..."}'
```

---

## 参考链接

- [APISIX jwt-auth 插件文档](https://apisix.apache.org/docs/apisix/plugins/jwt-auth/)
- [APISIX openid-connect 插件文档](https://apisix.apache.org/docs/apisix/plugins/openid-connect/)
- [Keycloak OIDC 配置指南](https://www.keycloak.org/docs/latest/securing_apps/)
