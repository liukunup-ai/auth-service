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

func (m *customUserModel) FindByUsername(username string) (*User, error) {
	var user User
	err := m.conn.QueryRow(&user, "SELECT * FROM users WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *customUserModel) FindByEmail(email string) (*User, error) {
	var user User
	err := m.conn.QueryRow(&user, "SELECT * FROM users WHERE email = ?", email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
