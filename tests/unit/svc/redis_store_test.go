package svc_test

import (
	"context"
	"testing"
	"time"

	"auth-service/internal/svc"

	"github.com/stretchr/testify/assert"

	"auth-service/tests/common")

func TestRedisStore(t *testing.T) {
	h := common.NewTestHelper(t)
	defer h.Cleanup()

	// Initialize service context with Redis
	h.SetupServiceContext(true)

	// Manually create RedisStore using the helper's redis client
	// Note: h.GetServiceContext().Redis is the client.

	ctx := context.Background()
	store := svc.NewRedisStore(ctx, h.GetServiceContext().Redis, "captcha:", 10*time.Minute)

	t.Run("Set and Get", func(t *testing.T) {
		err := store.Set("test_id", "1234")
		assert.NoError(t, err)

		// Get without clear
		val := store.Get("test_id", false)
		assert.Equal(t, "1234", val)

		// Verify stored value in redis directly
		rVal, err := h.GetServiceContext().Redis.Get(ctx, "captcha:test_id").Result()
		assert.NoError(t, err)
		assert.Equal(t, "1234", rVal)
	})

	t.Run("Get and Clear", func(t *testing.T) {
		err := store.Set("test_id_2", "5678")
		assert.NoError(t, err)

		// Get with clear
		val := store.Get("test_id_2", true)
		assert.Equal(t, "5678", val)

		// Should be gone
		val2 := store.Get("test_id_2", false)
		assert.Empty(t, val2)
	})

	t.Run("Verify", func(t *testing.T) {
		err := store.Set("test_id_3", "ABCd")
		assert.NoError(t, err)

		// Verify correct answer (case insensitive)
		ok := store.Verify("test_id_3", "abcd", false)
		assert.True(t, ok)

		// Verify incorrect answer
		ok = store.Verify("test_id_3", "wrong", false)
		assert.False(t, ok)

		// Verify with clear
		ok = store.Verify("test_id_3", "abcd", true)
		assert.True(t, ok)

		// Should be gone
		ok = store.Verify("test_id_3", "abcd", false)
		assert.False(t, ok)
	})

	t.Run("Set Error", func(t *testing.T) {
		// Impossible to simulate redis error with miniredis easily unless we close it,
		// but closing it might affect other tests.
		// Skip for now or try with a closed client if possible.
	})
}
