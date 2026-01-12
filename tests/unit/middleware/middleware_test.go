package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth-service/internal/middleware"
)

type mockClaims struct {
	UserID int64
}

func (c *mockClaims) GetUserID() int64 {
	return c.UserID
}
func (c *mockClaims) VerifyExpiresAt() bool {
	return true
}

func TestAuthInterceptorMiddleware_Handle(t *testing.T) {
	m := middleware.NewAuthInterceptorMiddleware()

	// Test 1: No token
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	m.Handle(nextHandler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 when no token, got %d", w.Code)
	}

	// Test 2: Valid token
	m.SetTokenValidator(func(tokenString string) (middleware.Claims, error) {
		if tokenString == "valid-token" {
			return &mockClaims{UserID: 123}, nil
		}
		return nil, errors.New("invalid token")
	})

	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w = httptest.NewRecorder()

	capturedUserID := int64(0)
	nextHandler = func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(middleware.UserIDKey)
		if id, ok := val.(int64); ok {
			capturedUserID = id
		}
	}

	m.Handle(nextHandler).ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200 OK, got %d body: %s", w.Code, w.Body.String())
	}
	if capturedUserID != 123 {
		t.Errorf("Expected UserID 123, got %d", capturedUserID)
	}

	// Test 3: Invalid token
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	m.Handle(nextHandler).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token, got %d", w.Code)
	}

	// Test 4: Token from Query
	req = httptest.NewRequest("GET", "/?token=valid-token", nil)
	w = httptest.NewRecorder()

	m.Handle(nextHandler).ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200 OK for query token, got %d", w.Code)
	}
}

type expiredClaims struct{}

func (c *expiredClaims) GetUserID() int64      { return 1 }
func (c *expiredClaims) VerifyExpiresAt() bool { return false }

func TestAuthInterceptorMiddleware_Expired(t *testing.T) {
	m := middleware.NewAuthInterceptorMiddleware()
	m.SetTokenValidator(func(tokenString string) (middleware.Claims, error) {
		return &expiredClaims{}, nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	w := httptest.NewRecorder()

	m.Handle(func(w http.ResponseWriter, r *http.Request) {}).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for expired token, got %d", w.Code)
	}
}
