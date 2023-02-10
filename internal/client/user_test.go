package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient_GetCurrentUser(t *testing.T) {
	t.Parallel()

	t.Run("retrieving the current user should deserialise correctly", func(t *testing.T) {
		client, _ := NewClient(testServerUrl, testUsername, testPassword)
		user, err := client.GetCurrentUser()

		expectedGroupMemberships := []GroupMembership{
			{Id: 1},
			{Id: 2},
		}

		assert.Nil(t, err)
		assert.Equal(t, "Example User", *user.CommonName)
		assert.NotEmpty(t, user.DateJoined)
		assert.Equal(t, "example@example.com", user.Email)
		assert.NotEmpty(t, *user.FirstLogin)
		assert.Equal(t, "Example", *user.FirstName)
		assert.Equal(t, false, user.GoogleAuth)
		assert.Equal(t, expectedGroupMemberships, user.GroupMemberships)
		assert.Equal(t, false, user.HasQuestionAndDashboard)
		assert.Equal(t, int64(1), user.Id)
		assert.Equal(t, true, user.IsActive)
		assert.Equal(t, true, user.IsInstaller)
		assert.Equal(t, true, user.IsQbnewb)
		assert.Equal(t, true, user.IsSuperuser)
		assert.NotEmpty(t, *user.LastLogin)
		assert.Equal(t, "User", *user.LastName)
		assert.Equal(t, false, user.LdapAuth)
		assert.Nil(t, user.Locale)
		assert.NotEmpty(t, *user.UpdatedAt)
	})
}

func TestClient_GetUser(t *testing.T) {
	t.Parallel()

	client, _ := NewClient(testServerUrl, testUsername, testPassword)

	t.Run("requesting a user that exists should return that user", func(t *testing.T) {
		user, err := client.GetUser(1)

		expectedGroupMemberships := []GroupMembership{
			{Id: 1},
			{Id: 2},
		}

		assert.Nil(t, err)
		assert.Equal(t, "Example User", *user.CommonName)
		assert.NotEmpty(t, user.DateJoined)
		assert.Equal(t, "example@example.com", user.Email)
		assert.Equal(t, "Example", *user.FirstName)
		assert.Equal(t, false, user.GoogleAuth)
		assert.Equal(t, int64(1), user.Id)
		assert.Equal(t, true, user.IsActive)
		assert.Equal(t, true, user.IsQbnewb)
		assert.Equal(t, true, user.IsSuperuser)
		assert.NotEmpty(t, *user.LastLogin)
		assert.Equal(t, "User", *user.LastName)
		assert.Equal(t, false, user.LdapAuth)
		assert.Nil(t, user.Locale)
		assert.NotEmpty(t, *user.UpdatedAt)
		assert.Equal(t, expectedGroupMemberships, user.GroupMemberships)

	})
	t.Run("requesting a user that doesn't exist should return an error", func(t *testing.T) {
		user, err := client.GetUser(2)

		assert.Nil(t, user)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}
