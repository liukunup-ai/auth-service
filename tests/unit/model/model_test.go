package model_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	model "auth-service/model/mysql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestUserModel_Insert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := model.NewUserModel(conn)

	user := &model.User{
		PublicId:      "pub_123",
		Nickname:      sql.NullString{String: "Nick", Valid: true},
		Username:      "user1",
		Email:         "email@test.com",
		PasswordHash:  "hash",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		AccountStatus: 1,
	}

	mock.ExpectExec("insert into `user`").
		WithArgs(user.PublicId, user.Nickname, user.Username, user.Email, user.EmailVerified, user.Phone, user.PhoneVerified, user.PasswordHash, user.PasswordSalt, user.MfaSecret, user.MfaEnabled, user.AccountStatus, user.FailedLoginAttempts, user.LockoutUntil, user.LastLoginAt, user.DeletedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = m.Insert(context.Background(), user)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}
}

func mockUserRows(id uint64, username, email string) *sqlmock.Rows {
	cols := []string{
		"id", "public_id", "nickname", "username", "email",
		"email_verified", "phone", "phone_verified", "password_hash", "password_salt",
		"mfa_secret", "mfa_enabled", "account_status", "failed_login_attempts",
		"lockout_until", "last_login_at", "created_at", "updated_at", "deleted_at",
	}
	return sqlmock.NewRows(cols).AddRow(
		id, "pub_"+username, sql.NullString{String: "Nick", Valid: true}, username, email,
		0, sql.NullString{String: "123456", Valid: true}, 0, "hash", sql.NullString{},
		sql.NullString{}, 0, 1, 0,
		sql.NullTime{}, sql.NullTime{}, time.Now(), time.Now(), sql.NullTime{},
	)
}

func TestUserModel_FindOne(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := model.NewUserModel(conn)

	rows := mockUserRows(1, "user1", "email@test.com")
	mock.ExpectQuery("select (.+) from `user` where `id` = ?").
		WithArgs(1).
		WillReturnRows(rows)

	res, err := m.FindOne(context.Background(), 1)
	if err != nil {
		t.Errorf("FindOne failed: %v", err)
	} else {
		if res.Username != "user1" {
			t.Errorf("Expected username user1, got %s", res.Username)
		}
	}

	// Test NotFound
	mock.ExpectQuery("select (.+) from `user` where `id` = ?").
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	_, err = m.FindOne(context.Background(), 99)
	if err != model.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestUserModel_FindOneByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := model.NewUserModel(conn)

	rows := mockUserRows(1, "user1", "email@test.com")
	mock.ExpectQuery("select (.+) from `user` where `username` = ?").
		WithArgs("user1").
		WillReturnRows(rows)

	res, err := m.FindOneByUsername(context.Background(), "user1")
	if err != nil {
		t.Errorf("FindOneByUsername failed: %v", err)
	} else {
		if res.Email != "email@test.com" {
			t.Errorf("Expected email email@test.com, got %s", res.Email)
		}
	}
}

func TestUserModel_FindOneByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := model.NewUserModel(conn)

	rows := mockUserRows(1, "user1", "email@test.com")
	mock.ExpectQuery("select (.+) from `user` where `email` = ?").
		WithArgs("email@test.com").
		WillReturnRows(rows)

	res, err := m.FindOneByEmail(context.Background(), "email@test.com")
	if err != nil {
		t.Errorf("FindOneByEmail failed: %v", err)
	} else {
		if res.Username != "user1" {
			t.Errorf("Expected username user1, got %s", res.Username)
		}
	}
}

func TestUserModel_FindOneByPhone(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := model.NewUserModel(conn)

	rows := mockUserRows(1, "user1", "email@test.com")
	mock.ExpectQuery("select (.+) from `user` where `phone` = ?").
		WithArgs("123456").
		WillReturnRows(rows)

	res, err := m.FindOneByPhone(context.Background(), "123456")
	if err != nil {
		t.Errorf("FindOneByPhone failed: %v", err)
	} else {
		if res.Phone.String != "123456" {
			t.Errorf("Expected phone 123456, got %s", res.Phone.String)
		}
	}
}

func TestUserModel_FindOneByPublicId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := model.NewUserModel(conn)

	rows := mockUserRows(1, "user1", "email@test.com")
	mock.ExpectQuery("select (.+) from `user` where `public_id` = ?").
		WithArgs("get_pub").
		WillReturnRows(rows)

	res, err := m.FindOneByPublicId(context.Background(), "get_pub")
	if err != nil {
		t.Errorf("FindOneByPublicId failed: %v", err)
	} else {
		if res.PublicId != "pub_user1" {
			t.Errorf("Expected pub_user1, got %s", res.PublicId)
		}
	}
}

func TestUserModel_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := model.NewUserModel(conn)

	user := &model.User{
		Id:            1,
		PublicId:      "pub_1",
		Username:      "user1",
		Email:         "email",
		AccountStatus: 1,
	}

	mock.ExpectExec("update `user`").
		WithArgs(user.PublicId, user.Nickname, user.Username, user.Email, user.EmailVerified, user.Phone, user.PhoneVerified, user.PasswordHash, user.PasswordSalt, user.MfaSecret, user.MfaEnabled, user.AccountStatus, user.FailedLoginAttempts, user.LockoutUntil, user.LastLoginAt, user.DeletedAt, user.Id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = m.Update(context.Background(), user)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
}

func TestUserModel_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := model.NewUserModel(conn)

	mock.ExpectExec("delete from `user`").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = m.Delete(context.Background(), 1)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}
