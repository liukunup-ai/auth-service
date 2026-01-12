package mysql

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		withSession(session sqlx.Session) UserModel
		FindOneByPhone(ctx context.Context, phone string) (*User, error)
	}
	customUserModel struct {
		*defaultUserModel
	}
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn),
	}
}

func (m *customUserModel) withSession(session sqlx.Session) UserModel {
	return NewUserModel(sqlx.NewSqlConnFromSession(session))
}

func (m *defaultUserModel) FindOneByPhone(ctx context.Context, phone string) (*User, error) {
	var resp User
	query := fmt.Sprintf("select %s from %s where `phone` = ? limit 1", userRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, phone)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// AccountStatus
const (
	UserStatusActive   = 1 // 用户状态: 正常
	UserStatusLocked   = 2 // 用户状态: 锁定
	UserStatusDisabled = 3 // 用户状态: 禁用
)

// generateSecureToken 生成安全的随机令牌
func generateSecureToken(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := make([]byte, length)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	// 将随机字节映射到字符集
	for i := range token {
		token[i] = charset[int(token[i])%len(charset)]
	}

	return string(token), nil
}
