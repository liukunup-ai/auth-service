package types

import "github.com/pkg/errors"

var (
	ErrCaptchaInvalid = errors.New("invalid captcha")
	ErrCaptchaRequired = errors.New("captcha is required")
	ErrUsernameTaken  = errors.New("username already exists")
	ErrEmailTaken     = errors.New("email already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrGenerateToken = errors.New("failed to generate token")
	ErrDatabaseError  = errors.New("database error")
)