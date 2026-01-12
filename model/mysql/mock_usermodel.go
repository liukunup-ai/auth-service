package mysql

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MockUserModel struct {
	InsertFunc            func(ctx context.Context, data *User) (sql.Result, error)
	FindOneFunc           func(ctx context.Context, id uint64) (*User, error)
	FindOneByEmailFunc    func(ctx context.Context, email string) (*User, error)
	FindOneByPublicIdFunc func(ctx context.Context, publicId string) (*User, error)
	FindOneByUsernameFunc func(ctx context.Context, username string) (*User, error)
	FindOneByPhoneFunc    func(ctx context.Context, phone string) (*User, error)
	UpdateFunc            func(ctx context.Context, data *User) error
	DeleteFunc            func(ctx context.Context, id uint64) error
	WithSessionFunc       func(session sqlx.Session) UserModel
}

func (m *MockUserModel) Insert(ctx context.Context, data *User) (sql.Result, error) {
	if m.InsertFunc != nil {
		return m.InsertFunc(ctx, data)
	}
	return nil, nil
}

func (m *MockUserModel) FindOne(ctx context.Context, id uint64) (*User, error) {
	if m.FindOneFunc != nil {
		return m.FindOneFunc(ctx, id)
	}
	return nil, ErrNotFound
}

func (m *MockUserModel) FindOneByEmail(ctx context.Context, email string) (*User, error) {
	if m.FindOneByEmailFunc != nil {
		return m.FindOneByEmailFunc(ctx, email)
	}
	return nil, ErrNotFound
}

func (m *MockUserModel) FindOneByPublicId(ctx context.Context, publicId string) (*User, error) {
	if m.FindOneByPublicIdFunc != nil {
		return m.FindOneByPublicIdFunc(ctx, publicId)
	}
	return nil, ErrNotFound
}

func (m *MockUserModel) FindOneByUsername(ctx context.Context, username string) (*User, error) {
	if m.FindOneByUsernameFunc != nil {
		return m.FindOneByUsernameFunc(ctx, username)
	}
	return nil, ErrNotFound
}

func (m *MockUserModel) FindOneByPhone(ctx context.Context, phone string) (*User, error) {
	if m.FindOneByPhoneFunc != nil {
		return m.FindOneByPhoneFunc(ctx, phone)
	}
	return nil, ErrNotFound
}

func (m *MockUserModel) Update(ctx context.Context, data *User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, data)
	}
	return nil
}

func (m *MockUserModel) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockUserModel) withSession(session sqlx.Session) UserModel {
	if m.WithSessionFunc != nil {
		return m.WithSessionFunc(session)
	}
	return m
}
