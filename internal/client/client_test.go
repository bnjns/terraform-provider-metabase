package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testServerUrl = "http://localhost:3000"
	testUsername  = "example@example.com"
	testPassword  = "password"
)

func TestNewClient(t *testing.T) {
	t.Run("providing no host should return error", func(t *testing.T) {
		c, err := NewClient("", testUsername, testPassword)

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "host", "Did not receive expected error")
	})

	t.Run("providing no username should return error", func(t *testing.T) {
		c, err := NewClient(testServerUrl, "", testPassword)

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "username", "Did not receive expected error")
	})

	t.Run("providing no password should return error", func(t *testing.T) {
		c, err := NewClient(testServerUrl, testUsername, "")

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "password", "Did not receive expected error")
	})

	t.Run("providing valid credentials should initialise a client", func(t *testing.T) {
		c, err := NewClient(testServerUrl, testUsername, testPassword)

		assert.Nil(t, err, "Expected no error")
		assert.NotNil(t, c, "Did not receive a valid client")
		assert.NotNilf(t, c.HttpClient, "Client does not have a valid HTTP client configured")
	})

	t.Run("providing valid credentials should sign in once initialised", func(t *testing.T) {
		c, _ := NewClient(testServerUrl, testUsername, testPassword)

		assert.NotEmpty(t, c.SessionId)
	})

	t.Run("providing invalid credentials should return an error", func(t *testing.T) {
		c, err := NewClient(testServerUrl, "bad", "credentials")

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "did not match stored password", "Did not receive expected error")
	})
}
