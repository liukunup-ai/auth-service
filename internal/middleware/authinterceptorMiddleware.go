package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	BearerPrefix = "Bearer "

	TokenHeaderKey = "Authorization"
	TokenQueryKey  = "token"

	TokenSourceHeader = "header"
	TokenSourceQuery  = "query"

	// 上下文键
	UserIDKey contextKey = "userID"
)

// contextKey 自定义上下文键类型
type contextKey string

// Claims 接口定义令牌声明需要的方法
type Claims interface {
	GetUserID() int64
	VerifyExpiresAt() bool
}

type TokenInfo struct {
	UserID int64
	Token  string
	Source string
}

// TokenValidator 令牌验证器函数类型
type TokenValidator func(tokenString string) (Claims, error)

type AuthInterceptorMiddleware struct {
	tokenValidator TokenValidator
}

// NewAuthInterceptorMiddleware 创建新的认证拦截器中间件
func NewAuthInterceptorMiddleware() *AuthInterceptorMiddleware {
	return &AuthInterceptorMiddleware{}
}

// SetTokenValidator 设置令牌验证器
func (m *AuthInterceptorMiddleware) SetTokenValidator(validator TokenValidator) {
	m.tokenValidator = validator
}

func (m *AuthInterceptorMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 根据配置获取 Token
		tokenInfo, err := m.extractToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// 2. 解析并验证 JWT 令牌
		if m.tokenValidator == nil {
			http.Error(w, "令牌验证器未设置", http.StatusInternalServerError)
			return
		}
		claims, err := m.tokenValidator(tokenInfo.Token)
		if err != nil {
			http.Error(w, fmt.Sprintf("无效的令牌: %v", err), http.StatusUnauthorized)
			return
		}

		// 3. 验证令牌是否过期
		if !claims.VerifyExpiresAt() {
			http.Error(w, "令牌已过期", http.StatusUnauthorized)
			return
		}

		// 4. 将用户信息和 Token 来源存入上下文
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.GetUserID())
		ctx = context.WithValue(ctx, contextKey("tokenSource"), tokenInfo.Source)
		ctx = context.WithValue(ctx, contextKey("rawToken"), tokenInfo.Token)

		newReq := r.WithContext(ctx)

		// 5. 调用下一个处理器
		next(w, newReq)
	}
}

func (m *AuthInterceptorMiddleware) extractToken(r *http.Request) (*TokenInfo, error) {
	// 优先从 Header 中获取
	if token, err := m.extractTokenFromHeader(r); err == nil {
		return token, nil
	}

	// 其次从 Query 参数中获取
	if token, err := m.extractTokenFromQuery(r); err == nil {
		return token, nil
	}

	return nil, errors.New("未找到有效的认证令牌")
}

func (m *AuthInterceptorMiddleware) extractTokenFromHeader(r *http.Request) (*TokenInfo, error) {
	authHeader := r.Header.Get(TokenHeaderKey)
	if authHeader == "" {
		return nil, errors.New("缺少 Authorization 头")
	}

	if !strings.HasPrefix(authHeader, BearerPrefix) {
		return nil, errors.New("Authorization 头格式错误，应为 'Bearer {token}'")
	}

	tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
	if tokenString == "" {
		return nil, errors.New("令牌不能为空")
	}

	return &TokenInfo{
		Token:  tokenString,
		Source: TokenSourceHeader,
	}, nil
}

func (m *AuthInterceptorMiddleware) extractTokenFromQuery(r *http.Request) (*TokenInfo, error) {
	tokenString := r.URL.Query().Get(TokenQueryKey)
	if tokenString == "" {
		return nil, errors.New("缺少 token 查询参数")
	}

	// 清理可能的空格
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, errors.New("token 查询参数不能为空")
	}

	return &TokenInfo{
		Token:  tokenString,
		Source: TokenSourceQuery,
	}, nil
}

// 移除parseToken方法，使用tokenValidator替代
