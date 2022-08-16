package client

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testUsername = "username"
	testPassword = "password"
)

func createMockServer(handlerFn func(writer http.ResponseWriter, request *http.Request) bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/session" && request.Method == "POST" {
			var authRequest AuthDetails
			defer request.Body.Close()
			body, _ := ioutil.ReadAll(request.Body)
			json.Unmarshal(body, &authRequest)

			if authRequest.Username == testUsername && authRequest.Password == testPassword {
				writer.WriteHeader(200)
				writer.Write([]byte(`{"id": "metabase-session-id"}`))
			} else {
				writer.WriteHeader(400)
				writer.Write([]byte(`{"errors": {"password": "did not match stored password"}`))
			}

			return
		}

		if !handlerFn(writer, request) {
			writer.WriteHeader(404)
			writer.Write([]byte(`{"error": "URL not found"}`))
		}
	}))
}

func TestNewClient(t *testing.T) {
	server := createMockServer(func(writer http.ResponseWriter, request *http.Request) bool {
		return false
	})
	defer server.Close()

	t.Run("providing no host should return error", func(t *testing.T) {
		c, err := NewClient("", testUsername, testPassword)

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "host", "Did not receive expected error")
	})

	t.Run("providing no username should return error", func(t *testing.T) {
		c, err := NewClient(server.URL, "", testPassword)

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "username", "Did not receive expected error")
	})

	t.Run("providing no password should return error", func(t *testing.T) {
		c, err := NewClient(server.URL, testUsername, "")

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "must provide", "Did not receive expected error")
		assert.ErrorContainsf(t, err, "password", "Did not receive expected error")
	})

	t.Run("providing valid credentials should initialise a client", func(t *testing.T) {
		c, err := NewClient(server.URL, testUsername, testPassword)

		assert.Nil(t, err, "Expected no error")
		assert.NotNil(t, c, "Did not receive a valid client")
		assert.NotNilf(t, c.HttpClient, "Client does not have a valid HTTP client configured")
	})

	t.Run("providing valid credentials should sign in once initialised", func(t *testing.T) {
		c, _ := NewClient(server.URL, testUsername, testPassword)

		assert.NotEmpty(t, c.SessionId)
	})

	t.Run("providing invalid credentials should return an error", func(t *testing.T) {
		c, err := NewClient(server.URL, "bad", "credentials")

		assert.Nil(t, c, "Expected nil client")
		assert.ErrorContainsf(t, err, "did not match stored password", "Did not receive expected error")
	})
}

func TestClient_doGet(t *testing.T) {
	server := createMockServer(func(writer http.ResponseWriter, request *http.Request) bool {
		if request.Method == "GET" && request.URL.Path == "/api/test" {
			writer.WriteHeader(200)
			writer.Header().Set("Content-Type", "application/json")
			writer.Write([]byte(`{"key": "value"}`))
			return true
		}

		return false
	})
	defer server.Close()

	client, _ := NewClient(server.URL, testUsername, testPassword)
	type TestResponse struct {
		Key string `json:"key"`
	}

	t.Run("sending a request to a valid URL should return a deserialised response", func(t *testing.T) {
		var response TestResponse
		err := client.doGet("/test", &response)

		assert.Nil(t, err)
		assert.Equal(t, "value", response.Key)
	})

	t.Run("sending a request to an invalid URL should return an error with the raw body", func(t *testing.T) {
		var response TestResponse
		err := client.doGet("/unknown", &response)

		assert.NotNil(t, err)
		assert.ErrorContains(t, err, `{"error": "URL not found"}`)
	})
}
