package svc_test

import (
	"testing"

	"auth-service/internal/svc"
)

func TestPasswordEncoder_Hash(t *testing.T) {
	encoder := &svc.PasswordEncoder{}
	password := "test123456"

	hash := encoder.Hash(password)

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should be different from original password")
	}

	// Hash should be consistent in length (bcrypt generates 60 character hashes)
	if len(hash) != 60 {
		t.Errorf("Expected hash length 60, got %d", len(hash))
	}
}

func TestPasswordEncoder_Compare(t *testing.T) {
	encoder := &svc.PasswordEncoder{}
	password := "test123456"

	hash := encoder.Hash(password)

	tests := []struct {
		name     string
		hash     string
		password string
		want     bool
	}{
		{
			name:     "正确的密码",
			hash:     hash,
			password: password,
			want:     true,
		},
		{
			name:     "错误的密码",
			hash:     hash,
			password: "wrongpassword",
			want:     false,
		},
		{
			name:     "空密码",
			hash:     hash,
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encoder.Compare(tt.hash, tt.password); got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordEncoder_HashDifferentPasswords(t *testing.T) {
	encoder := &svc.PasswordEncoder{}

	password1 := "password1"
	password2 := "password2"

	hash1 := encoder.Hash(password1)
	hash2 := encoder.Hash(password2)

	if hash1 == hash2 {
		t.Error("Different passwords should produce different hashes")
	}
}

func TestPasswordEncoder_HashSamePasswordTwice(t *testing.T) {
	encoder := &svc.PasswordEncoder{}

	password := "test123456"

	hash1 := encoder.Hash(password)
	hash2 := encoder.Hash(password)

	// bcrypt includes a random salt, so the same password should produce different hashes
	if hash1 == hash2 {
		t.Error("Same password hashed twice should produce different hashes due to random salt")
	}

	// But both should verify correctly
	if !encoder.Compare(hash1, password) {
		t.Error("First hash should verify correctly")
	}

	if !encoder.Compare(hash2, password) {
		t.Error("Second hash should verify correctly")
	}
}
