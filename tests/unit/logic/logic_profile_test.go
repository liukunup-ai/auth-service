package logic_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"auth-service/internal/logic"
	"auth-service/internal/types"
	model "auth-service/model/mysql"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetProfileLogic_GetProfile(t *testing.T) {
	// Setup service context
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()
	svcCtx := setupTestServiceContext(t, mock)

	// User to return
	testUser := &model.User{
		Id:        1, // Internal ID matches context
		PublicId:  "pub_123",
		Username:  "testuser",
		Email:     "test@example.com",
		Phone:     sql.NullString{String: "123456789", Valid: true},
		Nickname:  sql.NullString{String: "Test Nick", Valid: true},
		CreatedAt: time.Now(),
	}

	svcCtx.UserModel = &model.MockUserModel{
		FindOneFunc: func(ctx context.Context, id uint64) (*model.User, error) {
			if id == 1 {
				return testUser, nil
			}
			return nil, model.ErrNotFound
		},
	}

	// Context with userID
	ctx := context.WithValue(context.Background(), "userID", int64(1))

	l := logic.NewGetProfileLogic(ctx, svcCtx)
	resp, err := l.GetProfile()

	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	if resp.UserID != "pub_123" {
		t.Errorf("Expected UserID pub_123, got %s", resp.UserID)
	}
	if resp.Username != "testuser" {
		t.Errorf("Expected Username testuser, got %s", resp.Username)
	}
}

func TestChangePasswordLogic_ChangePassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()
	svcCtx := setupTestServiceContext(t, mock)

	oldHash := svcCtx.PasswordEncoder.Hash("oldPassword")

	testUser := &model.User{
		Id:           1,
		Username:     "testuser",
		PasswordHash: oldHash,
	}

	svcCtx.UserModel = &model.MockUserModel{
		FindOneFunc: func(ctx context.Context, id uint64) (*model.User, error) {
			if id == 1 {
				return testUser, nil
			}
			return nil, model.ErrNotFound
		},
		UpdateFunc: func(ctx context.Context, data *model.User) error {
			if data.Id != 1 {
				t.Fatalf("Update called with wrong ID")
			}
			// Verify password changed
			if data.PasswordHash == oldHash {
				t.Error("Password was not updated")
			}
			return nil
		},
	}

	ctx := context.WithValue(context.Background(), "userID", int64(1))
	l := logic.NewChangePasswordLogic(ctx, svcCtx)

	req := &types.ChangePasswordReq{
		OldPassword: "oldPassword",
		NewPassword: "newPassword123",
	}

	resp, err := l.ChangePassword(req)

	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}
	if resp.Code != 200 {
		t.Errorf("Expected code 200, got %d", resp.Code)
	}
}
