package client

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClient_GetPermissionsGroup(t *testing.T) {
	server := createMockServer(func(writer http.ResponseWriter, request *http.Request) bool {
		if request.Method == "GET" && request.URL.Path == "/api/permissions/group/3" {
			writer.WriteHeader(200)
			writer.Header().Set("Content-Type", "application.json")
			writer.Write([]byte(`
{
	"id": 3,
	"name": "Example Group"
}
`))
			return true
		}

		return false
	})
	defer server.Close()

	client, _ := NewClient(server.URL, testUsername, testPassword)

	t.Run("requesting a permissions group that exists should return that group", func(t *testing.T) {
		group, err := client.GetPermissionsGroup(3)

		assert.Nil(t, err)
		assert.Equal(t, int64(3), group.Id)
		assert.Equal(t, "Example Group", group.Name)
	})

	t.Run("requesting a permissions group that doesn't exist should return an error", func(t *testing.T) {
		group, err := client.GetPermissionsGroup(4)

		assert.Nil(t, group)
		assert.ErrorContains(t, err, "not found")
	})
}

func TestClient_CreatePermissionsGroup(t *testing.T) {
	server := createMockServer(func(writer http.ResponseWriter, request *http.Request) bool {
		if request.Method == "POST" && request.URL.Path == "/api/permissions/group" {
			writer.WriteHeader(200)
			writer.Header().Set("Content-Type", "application.json")
			writer.Write([]byte(`
{
	"id": 3,
	"name": "Example Group"
}
`))
			return true
		}

		return false
	})
	defer server.Close()

	client, _ := NewClient(server.URL, testUsername, testPassword)

	t.Run("creating a permissions group should return the group ID", func(t *testing.T) {
		createReq := PermissionsGroupRequest{
			Name: "Example Group",
		}
		groupId, err := client.CreatePermissionsGroup(createReq)

		assert.Nil(t, err)
		assert.Equal(t, int64(3), groupId)
	})
}

func TestClient_UpdatePermissionsGroup(t *testing.T) {
	server := createMockServer(func(writer http.ResponseWriter, request *http.Request) bool {
		if request.Method == "PUT" && request.URL.Path == "/api/permissions/group/3" {
			writer.WriteHeader(200)
			writer.Header().Set("Content-Type", "application.json")
			writer.Write([]byte(`
{
	"id": 3,
	"name": "Updated"
}
`))
			return true
		}

		return false
	})
	defer server.Close()

	client, _ := NewClient(server.URL, testUsername, testPassword)
	updateReq := PermissionsGroupRequest{
		Name: "Updated",
	}

	t.Run("updating a permissions group that exists should be handled successfully", func(t *testing.T) {
		err := client.UpdatePermissionsGroup(3, updateReq)

		assert.Nil(t, err)
	})

	t.Run("deleting a permissions group that doesn't exist should return an error", func(t *testing.T) {
		err := client.UpdatePermissionsGroup(4, updateReq)

		assert.ErrorContains(t, err, "not found")
	})
}

func TestClient_DeletePermissionsGroup(t *testing.T) {
	server := createMockServer(func(writer http.ResponseWriter, request *http.Request) bool {
		if request.Method == "DELETE" && request.URL.Path == "/api/permissions/group/3" {
			writer.WriteHeader(204)
			return true
		}

		return false
	})
	defer server.Close()

	client, _ := NewClient(server.URL, testUsername, testPassword)

	t.Run("deleting a permissions group that exists should be handled successfully", func(t *testing.T) {
		err := client.DeletePermissionsGroup(3)

		assert.Nil(t, err)
	})

	t.Run("deleting a permissions group that doesn't exist should return an error", func(t *testing.T) {
		err := client.DeletePermissionsGroup(4)

		assert.ErrorContains(t, err, "not found")
	})
}
