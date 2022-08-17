package client

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClient_GetCurrentUser(t *testing.T) {
	t.Run("retrieving the current user should deserialise correctly", func(t *testing.T) {
		server := createMockServer(func(writer http.ResponseWriter, request *http.Request) bool {
			if request.Method == "GET" && request.URL.Path == "/api/user/current" {
				writer.WriteHeader(200)
				writer.Header().Set("Content-Type", "application/json")
				writer.Write([]byte(`
{
    "common_name": "Example User",
    "date_joined": "2022-08-16T11:34:48.35",
    "email": "example@example.com",
    "first_login": "2022-08-16T11:35:23.233679+01:00",
    "first_name": "Example",
    "google_auth": false,
    "group_ids": [
        1,
        2
    ],
    "has_invited_second_user": false,
    "has_question_and_dashboard": false,
    "id": 1,
    "is_active": true,
    "is_installer": true,
    "is_qbnewb": true,
    "is_superuser": true,
    "last_login": "2022-08-16T11:34:48.516",
    "last_name": "User",
    "ldap_auth": false,
    "locale": null,
    "login_attributes": null,
    "personal_collection_id": 1,
    "sso_source": null,
    "updated_at": "2022-08-16T11:34:48.516"
}
`))
				return true
			}

			return false
		})
		defer server.Close()

		client, _ := NewClient(server.URL, testUsername, testPassword)
		user, err := client.GetCurrentUser()

		expectedGroupMemberships := []GroupMembership{
			{Id: 1},
			{Id: 2},
		}

		assert.Nil(t, err)
		assert.Equal(t, "Example User", *user.CommonName)
		assert.Equal(t, "2022-08-16T11:34:48.35", user.DateJoined)
		assert.Equal(t, "example@example.com", user.Email)
		assert.Equal(t, "2022-08-16T11:35:23.233679+01:00", *user.FirstLogin)
		assert.Equal(t, "Example", *user.FirstName)
		assert.Equal(t, false, user.GoogleAuth)
		assert.Equal(t, expectedGroupMemberships, user.GroupMemberships)
		assert.Equal(t, false, user.HasInvitedSecondUser)
		assert.Equal(t, false, user.HasQuestionAndDashboard)
		assert.Equal(t, int64(1), user.Id)
		assert.Equal(t, true, user.IsActive)
		assert.Equal(t, true, user.IsInstaller)
		assert.Equal(t, true, user.IsQbnewb)
		assert.Equal(t, true, user.IsSuperuser)
		assert.Equal(t, "2022-08-16T11:34:48.516", *user.LastLogin)
		assert.Equal(t, "User", *user.LastName)
		assert.Equal(t, false, user.LdapAuth)
		assert.Nil(t, user.Locale)
		assert.Equal(t, "2022-08-16T11:34:48.516", *user.UpdatedAt)

	})

	t.Run("retrieving the current user with all the optional fields missing should deserialise correctly", func(t *testing.T) {
		server := createMockServer(func(writer http.ResponseWriter, request *http.Request) bool {
			if request.Method == "GET" && request.URL.Path == "/api/user/current" {
				writer.WriteHeader(200)
				writer.Header().Set("Content-Type", "application/json")
				writer.Write([]byte(`
{
    "common_name": null,
    "date_joined": "2022-08-16T11:34:48.35",
    "email": "example@example.com",
    "first_login": null,
    "first_name": null,
    "google_auth": false,
    "group_ids": [],
    "has_invited_second_user": false,
    "has_question_and_dashboard": false,
    "id": 1,
    "is_active": true,
    "is_installer": true,
    "is_qbnewb": true,
    "is_superuser": true,
    "last_login": null,
    "last_name": null,
    "ldap_auth": false,
    "locale": null,
    "login_attributes": null,
    "personal_collection_id": 1,
    "sso_source": null,
    "updated_at": null
}
`))
				return true
			}

			return false
		})
		defer server.Close()

		client, _ := NewClient(server.URL, testUsername, testPassword)
		user, err := client.GetCurrentUser()

		assert.Nil(t, err)
		assert.Nil(t, user.CommonName)
		assert.Equal(t, "2022-08-16T11:34:48.35", user.DateJoined)
		assert.Equal(t, "example@example.com", user.Email)
		assert.Nil(t, user.FirstLogin)
		assert.Nil(t, user.FirstName)
		assert.Equal(t, false, user.GoogleAuth)
		assert.Equal(t, []GroupMembership{}, user.GroupMemberships)
		assert.Equal(t, false, user.HasInvitedSecondUser)
		assert.Equal(t, false, user.HasQuestionAndDashboard)
		assert.Equal(t, int64(1), user.Id)
		assert.Equal(t, true, user.IsActive)
		assert.Equal(t, true, user.IsInstaller)
		assert.Equal(t, true, user.IsQbnewb)
		assert.Equal(t, true, user.IsSuperuser)
		assert.Nil(t, user.LastLogin)
		assert.Nil(t, user.LastName)
		assert.Equal(t, false, user.LdapAuth)
		assert.Nil(t, user.Locale)
		assert.Nil(t, user.UpdatedAt)

	})
}

func TestClient_GetUser(t *testing.T) {
	server := createMockServer(func(writer http.ResponseWriter, request *http.Request) bool {
		if request.Method == "GET" && request.URL.Path == "/api/user/1" {
			writer.WriteHeader(200)
			writer.Header().Set("Content-Type", "application/json")
			writer.Write([]byte(`
{
    "common_name": "Example User",
    "date_joined": "2022-08-16T17:32:42.453",
    "email": "example@example.com",
    "first_name": "Example",
    "google_auth": false,
    "id": 1,
    "is_active": true,
    "is_qbnewb": true,
    "is_superuser": true,
    "last_login": "2022-08-16T17:32:42.522",
    "last_name": "User",
    "ldap_auth": false,
    "locale": null,
    "login_attributes": null,
    "sso_source": null,
    "updated_at": "2022-08-16T17:32:42.522",
    "user_group_memberships": [
        {
            "id": 1
        },
        {
            "id": 2
        }
    ]
}
`))
			return true
		} else if request.Method == "GET" && request.URL.Path == "/api/user/2" {
			writer.WriteHeader(404)
			writer.Header().Set("Content-Type", "application/json")
			writer.Write([]byte("Not found."))
			return true
		} else if request.Method == "GET" && request.URL.Path == "/api/user/3" {
			writer.WriteHeader(200)
			writer.Header().Set("Content-Type", "application/json")
			writer.Write([]byte(`
{
    "common_name": null,
    "date_joined": "2022-08-16T17:32:42.453",
    "email": "example@example.com",
    "first_name": null,
    "google_auth": false,
    "id": 1,
    "is_active": true,
    "is_qbnewb": true,
    "is_superuser": true,
    "last_login": null,
    "last_name": null,
    "ldap_auth": false,
    "locale": null,
    "login_attributes": null,
    "sso_source": null,
    "updated_at": null,
    "user_group_memberships": []
}
`))
			return true
		}

		return false
	})
	defer server.Close()

	client, _ := NewClient(server.URL, testUsername, testPassword)

	t.Run("requesting a user that exists should return that user", func(t *testing.T) {
		user, err := client.GetUser(1)

		expectedGroupMemberships := []GroupMembership{
			{Id: 1},
			{Id: 2},
		}

		assert.Nil(t, err)
		assert.Equal(t, "Example User", *user.CommonName)
		assert.Equal(t, "2022-08-16T17:32:42.453", user.DateJoined)
		assert.Equal(t, "example@example.com", user.Email)
		assert.Equal(t, "Example", *user.FirstName)
		assert.Equal(t, false, user.GoogleAuth)
		assert.Equal(t, int64(1), user.Id)
		assert.Equal(t, true, user.IsActive)
		assert.Equal(t, true, user.IsQbnewb)
		assert.Equal(t, true, user.IsSuperuser)
		assert.Equal(t, "2022-08-16T17:32:42.522", *user.LastLogin)
		assert.Equal(t, "User", *user.LastName)
		assert.Equal(t, false, user.LdapAuth)
		assert.Nil(t, user.Locale)
		assert.Equal(t, "2022-08-16T17:32:42.522", *user.UpdatedAt)
		assert.Equal(t, expectedGroupMemberships, user.GroupMemberships)

	})
	t.Run("requesting a user that doesn't exist should return an error", func(t *testing.T) {
		user, err := client.GetUser(2)

		assert.Nil(t, user)
		assert.NotNil(t, err)
		assert.ErrorContains(t, err, "Not found.")
	})
	t.Run("requesting a user with minimal properties should return that user", func(t *testing.T) {
		user, err := client.GetUser(3)

		assert.Nil(t, err)
		assert.Nil(t, user.CommonName)
		assert.Nil(t, user.FirstName)
		assert.Nil(t, user.LastLogin)
		assert.Nil(t, user.LastName)
		assert.Equal(t, []GroupMembership{}, user.GroupMemberships)
	})
}
