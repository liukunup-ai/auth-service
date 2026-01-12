package svc_test

import (
	"auth-service/internal/svc"
	"net/smtp"
	"testing"
)

func TestEmailClient_SendResetEmail(t *testing.T) {
	// Mock sendMail
	called := false
	mockSendMail := func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		called = true
		if from != "admin@example.com" {
			t.Errorf("Expected from admin@example.com, got %s", from)
		}
		if to[0] != "user@example.com" {
			t.Errorf("Expected to user@example.com, got %s", to[0])
		}
		return nil
	}

	svc.SetSendMailFunc(mockSendMail)
	defer svc.SetSendMailFunc(nil) // Restore (need to export access or just use internal test)

	config := &svc.Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "password",
		From:     "admin@example.com",
	}
	client := svc.NewClient(config)

	err := client.SendResetEmail("user@example.com", "http://reset.link")
	if err != nil {
		t.Errorf("SendResetEmail failed: %v", err)
	}
	if !called {
		t.Error("sendMail was not called")
	}
}
