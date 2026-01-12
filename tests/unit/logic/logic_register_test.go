package logic_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/logic"
	"auth-service/internal/svc"
	"auth-service/internal/types"
	model "auth-service/model/mysql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sony/sonyflake"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func setupTestServiceContext(t *testing.T, mock sqlmock.Sqlmock) *svc.ServiceContext {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	conn := sqlx.NewSqlConnFromDB(db)

	cfg := config.Config{
		Captcha: struct {
			Enable      bool
			ExpiresIn   int64
			Length      int
			CachePrefix string
		}{
			Enable:      false, // 测试时禁用验证码
			ExpiresIn:   300,
			Length:      6,
			CachePrefix: "test:captcha:",
		},
	}

	return &svc.ServiceContext{
		Config:          cfg,
		DB:              conn,
		PasswordEncoder: &svc.PasswordEncoder{},
		UserModel:       model.NewUserModel(conn),
	}
}

func TestRegisterLogic_Register_Success(t *testing.T) {
	// Setup service context with mock DB (needed for initialization but not used for user queries)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	svcCtx := setupTestServiceContext(t, mock)

	// Create manual mock for UserModel
	mockUserModel := &model.MockUserModel{
		FindOneByUsernameFunc: func(ctx context.Context, username string) (*model.User, error) {
			return nil, model.ErrNotFound
		},
		FindOneByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return nil, model.ErrNotFound
		},
		FindOneByPhoneFunc: func(ctx context.Context, phone string) (*model.User, error) {
			return nil, model.ErrNotFound
		},
		InsertFunc: func(ctx context.Context, data *model.User) (sql.Result, error) {
			return sqlmock.NewResult(1, 1), nil
		},
	}
	svcCtx.UserModel = mockUserModel

	// 创建 Sonyflake 用于生成 PublicId
	st := sonyflake.Settings{
		StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		MachineID: func() (uint16, error) {
			return 1, nil
		},
	}
	sf := sonyflake.NewSonyflake(st)
	if sf == nil {
		t.Fatal("Failed to create Sonyflake")
	}
	svcCtx.Sonyflake = sf

	l := logic.NewRegisterLogic(context.Background(), svcCtx)

	req := &types.RegisterReq{
		Username: "newuser",
		Password: "password123",
		Email:    "newuser@example.com",
		Phone:    "13800138000",
		Nickname: "New User",
	}

	resp, err := l.Register(req)

	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if resp != nil {
		if resp.Username != req.Username {
			t.Errorf("Username = %v, want %v", resp.Username, req.Username)
		}

		if resp.Email != req.Email {
			t.Errorf("Email = %v, want %v", resp.Email, req.Email)
		}
	}
}

func TestRegisterLogic_Register_UsernameTaken(t *testing.T) {
	// Setup service context with mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	svcCtx := setupTestServiceContext(t, mock)

	// Create manual mock for UserModel
	mockUserModel := &model.MockUserModel{
		FindOneByUsernameFunc: func(ctx context.Context, username string) (*model.User, error) {
			// Simulate existing user
			return &model.User{
				Id:       1,
				Username: "existinguser",
				Email:    "existing@example.com",
			}, nil
		},
		FindOneByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return nil, model.ErrNotFound
		},
		FindOneByPhoneFunc: func(ctx context.Context, phone string) (*model.User, error) {
			return nil, model.ErrNotFound
		},
	}
	svcCtx.UserModel = mockUserModel

	// 创建 Sonyflake
	st := sonyflake.Settings{
		StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		MachineID: func() (uint16, error) {
			return 1, nil
		},
	}
	sf := sonyflake.NewSonyflake(st)
	if sf == nil {
		t.Fatal("Failed to create Sonyflake")
	}
	svcCtx.Sonyflake = sf

	l := logic.NewRegisterLogic(context.Background(), svcCtx)

	req := &types.RegisterReq{
		Username: "existinguser",
		Password: "password123",
		Email:    "new@example.com",
	}

	_, err = l.Register(req)

	if err != types.ErrUsernameTaken {
		t.Errorf("Expected ErrUsernameTaken, got %v", err)
	}
}
