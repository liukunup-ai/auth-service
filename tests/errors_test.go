package tests

import (
	"testing"

	"auth-service/internal/types"
)

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrCaptchaInvalid",
			err:  types.ErrCaptchaInvalid,
			want: "invalid captcha",
		},
		{
			name: "ErrCaptchaRequired",
			err:  types.ErrCaptchaRequired,
			want: "captcha is required",
		},
		{
			name: "ErrUsernameTaken",
			err:  types.ErrUsernameTaken,
			want: "username already exists",
		},
		{
			name: "ErrEmailTaken",
			err:  types.ErrEmailTaken,
			want: "email already exists",
		},
		{
			name: "ErrPhoneTaken",
			err:  types.ErrPhoneTaken,
			want: "phone number already exists",
		},
		{
			name: "ErrUserNotFound",
			err:  types.ErrUserNotFound,
			want: "user not found",
		},
		{
			name: "ErrInvalidPassword",
			err:  types.ErrInvalidPassword,
			want: "invalid password",
		},
		{
			name: "ErrGenerateToken",
			err:  types.ErrGenerateToken,
			want: "failed to generate token",
		},
		{
			name: "ErrDatabaseError",
			err:  types.ErrDatabaseError,
			want: "database error",
		},
		{
			name: "ErrInvalidRefreshToken",
			err:  types.ErrInvalidRefreshToken,
			want: "invalid refresh token",
		},
		{
			name: "ErrUnauthorized",
			err:  types.ErrUnauthorized,
			want: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error message = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorEquality(t *testing.T) {
	// 测试错误相等性
	if types.ErrUserNotFound != types.ErrUserNotFound {
		t.Error("Same error should be equal")
	}

	if types.ErrUserNotFound == types.ErrInvalidPassword {
		t.Error("Different errors should not be equal")
	}
}
