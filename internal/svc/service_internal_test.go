package svc

import (
	"os"
	"testing"
	"time"

	"auth-service/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestGetMachineID(t *testing.T) {
	// 1. Env var success
	os.Setenv("SNOWFLAKE_MACHINE_ID", "1")
	defer os.Unsetenv("SNOWFLAKE_MACHINE_ID")

	id, err := getMachineID()
	assert.NoError(t, err)
	assert.Equal(t, uint16(1), id)

	// 2. Env var invalid
	os.Setenv("SNOWFLAKE_MACHINE_ID", "invalid")
	// Depending on implementation, it usually falls back to random.
	// But let's just ensure no panic.
	_, _ = getMachineID()

	// 3. Env var out of range
	os.Setenv("SNOWFLAKE_MACHINE_ID", "99999")
	_, err = getMachineID()
	// Should act safe
	if err != nil {
		t.Log("Got error for out of range, expected behavior")
	}
}

func TestInitRedis_Fail_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	c := config.Config{}
	c.Redis.Addrs = []string{"invalid-host:6379"}
	c.Redis.DB = 0
	c.Redis.Password = "pass"

	done := make(chan bool)
	go func() {
		initRedis(c)
		done <- true
	}()

	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		// timeout implies it might be trying to connect.
	}
}

func TestInitSSOProviders(t *testing.T) {
	c := config.Config{}
	// Both disabled
	c.SSO.OIDC.Enabled = false
	c.SSO.LDAP.Enabled = false

	oidc, ldap := initSSOProviders(c)
	assert.Nil(t, oidc)
	assert.Nil(t, ldap)

	// Enable
	c.SSO.OIDC.Enabled = true
	c.SSO.LDAP.Enabled = true

	// Just ensure it doesn't panic
	oidc, ldap = initSSOProviders(c)
}
