package svc

import "golang.org/x/crypto/bcrypt"

type PasswordEncoder struct{}

func (p *PasswordEncoder) Hash(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed)
}

func (p *PasswordEncoder) Compare(hashed, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}