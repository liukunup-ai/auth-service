package types

import "github.com/pkg/errors"

var (
	ErrCaptchaInvalid = errors.New("invalid captcha")
	ErrUsernameTaken  = errors.New("username already exists")
	ErrEmailTaken     = errors.New("email already exists")
	ErrDatabaseError  = errors.New("database error")
)