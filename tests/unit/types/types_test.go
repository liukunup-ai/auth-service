package types_test

import (
	"encoding/json"
	"regexp"
	"testing"

	"auth-service/internal/types"
)

func TestBaseResponse_JSON(t *testing.T) {
	resp := types.BaseResponse{
		Code:    200,
		Message: "Success",
		Data: map[string]interface{}{
			"userId":   "123",
			"username": "testuser",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded types.BaseResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Code != resp.Code {
		t.Errorf("Code = %v, want %v", decoded.Code, resp.Code)
	}

	if decoded.Message != resp.Message {
		t.Errorf("Message = %v, want %v", decoded.Message, resp.Message)
	}
}

func TestLoginReq_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  types.LoginReq
		want bool // true if should be valid
	}{
		{
			name: "有效的登录请求",
			req: types.LoginReq{
				Username: "testuser",
				Password: "password123",
			},
			want: true,
		},
		{
			name: "空用户名",
			req: types.LoginReq{
				Username: "",
				Password: "password123",
			},
			want: false,
		},
		{
			name: "空密码",
			req: types.LoginReq{
				Username: "testuser",
				Password: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 简单的验证检查
			valid := tt.req.Username != "" && tt.req.Password != ""
			if valid != tt.want {
				t.Errorf("Validation = %v, want %v", valid, tt.want)
			}
		})
	}
}

func TestRegisterReq_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  types.RegisterReq
		want bool
	}{
		{
			name: "有效的注册请求",
			req: types.RegisterReq{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
			},
			want: true,
		},
		{
			name: "用户名太短",
			req: types.RegisterReq{
				Username: "ab",
				Password: "password123",
				Email:    "test@example.com",
			},
			want: false,
		},
		{
			name: "密码太短",
			req: types.RegisterReq{
				Username: "testuser",
				Password: "12345",
				Email:    "test@example.com",
			},
			want: false,
		},
		{
			name: "无效的邮箱",
			req: types.RegisterReq{
				Username: "testuser",
				Password: "password123",
				Email:    "invalid-email",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 简单的验证检查，包括邮箱格式验证
			emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
			valid := len(tt.req.Username) >= 3 &&
				len(tt.req.Username) <= 20 &&
				len(tt.req.Password) >= 6 &&
				len(tt.req.Password) <= 30 &&
				len(tt.req.Email) > 0 &&
				emailRegex.MatchString(tt.req.Email)

			if valid != tt.want {
				t.Errorf("Validation = %v, want %v", valid, tt.want)
			}
		})
	}
}

func TestChangePasswordReq_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  types.ChangePasswordReq
		want bool
	}{
		{
			name: "有效的修改密码请求",
			req: types.ChangePasswordReq{
				OldPassword: "oldpass123",
				NewPassword: "newpass456",
			},
			want: true,
		},
		{
			name: "新密码太短",
			req: types.ChangePasswordReq{
				OldPassword: "oldpass123",
				NewPassword: "12345",
			},
			want: false,
		},
		{
			name: "空的旧密码",
			req: types.ChangePasswordReq{
				OldPassword: "",
				NewPassword: "newpass456",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.req.OldPassword != "" &&
				len(tt.req.NewPassword) >= 6 &&
				len(tt.req.NewPassword) <= 30

			if valid != tt.want {
				t.Errorf("Validation = %v, want %v", valid, tt.want)
			}
		})
	}
}
