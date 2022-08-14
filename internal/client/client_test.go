package client

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func getTestClientConfig() (string, AuthDetails) {
	host := os.Getenv("METABASE_HOST")
	username := os.Getenv("METABASE_USERNAME")
	password := os.Getenv("METABASE_PASSWORD")

	return host, AuthDetails{
		Username: username,
		Password: password,
	}
}

func TestNewClient(t *testing.T) {
	host, credentials := getTestClientConfig()

	t.Run("providing no host should return error", func(t *testing.T) {
		c, err := NewClient("", credentials.Username, credentials.Password)

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "host", "Did not receive expected error")
	})

	t.Run("providing no username should return error", func(t *testing.T) {
		c, err := NewClient(host, "", credentials.Password)

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "username", "Did not receive expected error")
	})

	t.Run("providing no password should return error", func(t *testing.T) {
		c, err := NewClient(host, credentials.Username, "")

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "password", "Did not receive expected error")
	})

	t.Run("providing valid credentials should initialise a client", func(t *testing.T) {
		c, err := NewClient(host, credentials.Username, credentials.Password)

		assert.Nil(t, err, "Expected no error")
		assert.NotNil(t, c, "Did not receive a valid client")
		assert.NotNilf(t, c.HttpClient, "Client does not have a valid HTTP client configured")
	})

	t.Run("providing valid credentials should sign in once initialised", func(t *testing.T) {
		c, _ := NewClient(host, credentials.Username, credentials.Password)

		assert.NotEmpty(t, c.SessionId)
	})

	t.Run("providing invalid credentials should return an error", func(t *testing.T) {
		c, err := NewClient(host, "bad", "credentials")

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "did not match stored password", "Did not receive expected error")
	})
}
