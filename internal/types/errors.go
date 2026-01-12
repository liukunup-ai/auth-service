package types

import "github.com/pkg/errors"

var (
	ErrCaptchaInvalid      = errors.New("invalid captcha")
	ErrCaptchaRequired     = errors.New("captcha is required")
	ErrUsernameTaken       = errors.New("username already exists")
	ErrEmailTaken          = errors.New("email already exists")
	ErrPhoneTaken          = errors.New("phone number already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrGenerateToken       = errors.New("failed to generate token")
	ErrDatabaseError       = errors.New("database error")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidResetToken   = errors.New("invalid reset token")
	ErrResetTokenExpired   = errors.New("reset token expired")
	ErrUnauthorized        = errors.New("unauthorized")

	// SSO 相关错误
	ErrSSOProviderNotEnabled = errors.New("SSO provider is not enabled")
	ErrInvalidState          = errors.New("invalid or expired state parameter")
	ErrSSOExchangeFailed     = errors.New("failed to exchange authorization code")
	ErrSSOGetUserInfoFailed  = errors.New("failed to get user info from SSO provider")
	ErrSSOResponseInvalid    = errors.New("invalid SSO response")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInternalServer        = errors.New("internal server error")
)
