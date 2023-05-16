package client

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testServerUrl = "http://localhost:3000"
	testUsername  = "example@example.com"
	testPassword  = "password"
)

var (
	testHeaders = map[string]string{}
)

func TestNewClient(t *testing.T) {
	t.Run("providing no host should return error", func(t *testing.T) {
		c, err := NewClient("", testUsername, testPassword, testHeaders)

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "host", "Did not receive expected error")
	})

	t.Run("providing no username should return error", func(t *testing.T) {
		c, err := NewClient(testServerUrl, "", testPassword, testHeaders)

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "username", "Did not receive expected error")
	})

	t.Run("providing no password should return error", func(t *testing.T) {
		c, err := NewClient(testServerUrl, testUsername, "", testHeaders)

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "password", "Did not receive expected error")
	})

	t.Run("providing valid credentials should initialise a client", func(t *testing.T) {
		c, err := NewClient(testServerUrl, testUsername, testPassword, testHeaders)

		assert.Nil(t, err, "Expected no error")
		assert.NotNil(t, c, "Did not receive a valid client")
		assert.NotNilf(t, c.HttpClient, "Client does not have a valid HTTP client configured")
	})

	t.Run("providing valid credentials should sign in once initialised", func(t *testing.T) {
		c, _ := NewClient(testServerUrl, testUsername, testPassword, testHeaders)

		assert.NotEmpty(t, c.SessionId)
	})

	t.Run("providing invalid credentials should return an error", func(t *testing.T) {
		c, err := NewClient(testServerUrl, "bad", "credentials", testHeaders)

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "did not match stored password", "Did not receive expected error")
	})

	t.Run("providing headers should be included in the client config", func(t *testing.T) {
		c, err := NewClient(testServerUrl, testUsername, testPassword, map[string]string{
			"Example": "Header",
		})

		assert.NoError(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, 1, len(c.Headers))
		assert.Equal(t, "Header", c.Headers["Example"])
	})
}

func TestNewClient_Headers(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/session", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id": "session-id"}`))
	})
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorizer") != "Bearer token" {
			w.WriteHeader(500)
			w.Write([]byte(`{}`))
			return
		}

		if req.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(500)
			w.Write([]byte(`{}`))
			return
		}

		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Run("headers should be added to requests", func(t *testing.T) {
		c, err := NewClient(server.URL, testUsername, testPassword, map[string]string{
			"Authorizer":   "Bearer token",
			"Content-Type": "invalid/type",
		})

		assert.NoError(t, err)
		assert.NotNil(t, c)

		err = c.doGet("/test", nil)
		assert.NoError(t, err)
	})
}
