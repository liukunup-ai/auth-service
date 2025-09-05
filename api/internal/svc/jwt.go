package svc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"auth-service/api/internal/config"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// 令牌类型
type TokenType string

// 令牌枚举值
const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// 令牌对
type TokenPair struct {
	AccessToken   string `json:"accessToken"`
	AccessExpire  int64  `json:"accessExpire"`
	RefreshToken  string `json:"refreshToken"`
	RefreshExpire int64  `json:"refreshExpire"`
}

type CustomClaims struct {
	jwt.StandardClaims

	UserID   uint64    `json:"userId"`
	Username string    `json:"username"`
	TokenID  string    `json:"tokenId"`
	Type     TokenType `json:"type"`
}

type JWT struct {
	accessSecret    []byte
	accessExpire    time.Duration
	refreshSecret   []byte
	refreshExpire   time.Duration
	rdb             redis.UniversalClient
	blacklistPrefix string
	mu              sync.Mutex
}

func NewJWT(c config.Config, rdb redis.UniversalClient) *JWT {
	return &JWT{
		accessSecret:    []byte(c.Auth.AccessSecret),
		accessExpire:    time.Duration(c.Auth.AccessExpire) * time.Second,
		refreshSecret:   []byte(c.Auth.RefreshSecret),
		refreshExpire:   time.Duration(c.Auth.RefreshExpire) * time.Second,
		rdb:             rdb,
		blacklistPrefix: c.Auth.BlacklistCachePrefix,
		mu:              sync.Mutex{},
	}
}

func (j *JWT) Generate(userID uint64, username string) (*TokenPair, error) {
	tokenID := generateTokenID()

	// 生成 Access Token
	accessToken, accessExpire, err := j.generateToken(userID, username, tokenID, AccessToken)
	if err != nil {
		return nil, err
	}

	// 生成 Refresh Token
	refreshToken, refreshExpire, err := j.generateToken(userID, username, tokenID, RefreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:   accessToken,
		AccessExpire:  accessExpire,
		RefreshToken:  refreshToken,
		RefreshExpire: refreshExpire,
	}, nil
}

func (j *JWT) generateToken(userID uint64, username string, tokenID string, tokenType TokenType) (string, int64, error) {

	var expireTime time.Time
	var secret []byte

	now := time.Now()

	switch tokenType {
	case AccessToken:
		expireTime = now.Add(j.accessExpire)
		secret = j.accessSecret
	case RefreshToken:
		expireTime = now.Add(j.refreshExpire)
		secret = j.refreshSecret
	default:
		return "", 0, errors.New("unsupported token type")
	}

	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		TokenID:  tokenID,
		Type:     tokenType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			Issuer:    "auth-service",
			Subject:   fmt.Sprintf("%d", userID),
			Id:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expireTime.Unix(), nil
}

func (j *JWT) VerifyAccessToken(tokenString string) (*CustomClaims, error) {
	return j.verifyToken(tokenString, AccessToken, j.accessSecret)
}

func (j *JWT) VerifyRefreshToken(tokenString string) (*CustomClaims, error) {
	return j.verifyToken(tokenString, RefreshToken, j.refreshSecret)
}

func (j *JWT) verifyToken(tokenString string, expectedType TokenType, secret []byte) (*CustomClaims, error) {
	// 检查 Token 是否在黑名单中
	if j.rdb != nil {
		isBlacklisted, err := j.isTokenBlacklisted(tokenString)
		if err != nil {
			return nil, fmt.Errorf("failed to check token blacklist: %v", err)
		}
		if isBlacklisted {
			return nil, errors.New("token is blacklisted")
		}
	}

	// 解析并验证 Token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// 检查Token类型
		if claims.Type != expectedType {
			return nil, errors.New("invalid token type")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (j *JWT) Refresh(refreshTokenString string) (*TokenPair, error) {
	// 验证 Refresh Token
	claims, err := j.VerifyRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %v", err)
	}

	// 将旧的 Token 加入黑名单
	if j.rdb != nil {
		// 计算剩余过期时间
		now := time.Now()
		refreshExpire := time.Unix(claims.ExpiresAt, 0).Sub(now)
		// 加入黑名单
		if err := j.BlacklistToken(refreshTokenString, refreshExpire); err != nil {
			return nil, fmt.Errorf("failed to blacklist old refresh token: %v", err)
		}
	}

	// 生成新的 Token 对
	return j.Generate(claims.UserID, claims.Username)
}

func (j *JWT) Logout(accessToken string, refreshToken string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.rdb == nil {
		return errors.New("rdb is not initialized")
	}

	// 验证并获取 Claims
	accessClaims, err := j.VerifyAccessToken(accessToken)
	if err != nil {
		return fmt.Errorf("failed to verify access token: %v", err)
	}
	refreshClaims, err := j.VerifyRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("failed to verify refresh token: %v", err)
	}

	// 计算剩余过期时间
	now := time.Now()
	accessExpire := time.Unix(accessClaims.ExpiresAt, 0).Sub(now)
	refreshExpire := time.Unix(refreshClaims.ExpiresAt, 0).Sub(now)

	// 将 Token 加入黑名单
	if err := j.BlacklistToken(accessToken, accessExpire); err != nil {
		return err
	}
	if err := j.BlacklistToken(refreshToken, refreshExpire); err != nil {
		return err
	}

	return nil
}

// 将 Token 加入到黑名单中
func (j *JWT) BlacklistToken(tokenString string, expire time.Duration) error {
	if j.rdb == nil {
		return errors.New("rdb is not initialized")
	}

	key := fmt.Sprintf("%s:%s", j.blacklistPrefix, tokenString)

	ctx := context.Background()
	return j.rdb.SetEx(ctx, key, "1", expire).Err()
}

func (j *JWT) isTokenBlacklisted(tokenString string) (bool, error) {
	key := fmt.Sprintf("%s%s", j.blacklistPrefix, tokenString)

	ctx := context.Background()
	exists, err := j.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func generateTokenID() string {
	return uuid.New().String()
}
