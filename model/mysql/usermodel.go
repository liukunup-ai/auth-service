package mysql

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		withSession(session sqlx.Session) UserModel
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

// AccountStatus
const (
	UserStatusActive   = 1 // 用户状态: 正常
	UserStatusLocked   = 2 // 用户状态: 锁定
	UserStatusDisabled = 3 // 用户状态: 禁用
)
