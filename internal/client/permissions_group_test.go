package client

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient_PermissionsGroup(t *testing.T) {
	client, _ := NewClient(testServerUrl, testUsername, testPassword)

	newGroupName := acctest.RandString(10)
	updatedGroupName := acctest.RandString(11)

	t.Run("you should be able to create a valid permissions group", func(t *testing.T) {
		groupId, err := client.CreatePermissionsGroup(PermissionsGroupRequest{
			Name: newGroupName,
		})

		assert.Nil(t, err)
		assert.NotZero(t, groupId)

		t.Run("you should be able to fetch the permission group", func(t *testing.T) {
			group, err := client.GetPermissionsGroup(groupId)

			assert.NoError(t, err)
			assert.NotZero(t, groupId)
			assert.Equal(t, newGroupName, group.Name)
		})

		t.Run("you should be able to update the permission group", func(t *testing.T) {
			err := client.UpdatePermissionsGroup(groupId, PermissionsGroupRequest{
				Name: updatedGroupName,
			})

			assert.Nil(t, err)
		})

		t.Run("you should be able to delete the permissions group", func(t *testing.T) {
			err := client.DeletePermissionsGroup(groupId)

			assert.NoError(t, err)
		})
	})

	t.Run("requesting a permissions group that doesn't exist should return an error", func(t *testing.T) {
		group, err := client.GetPermissionsGroup(1000)

		assert.Nil(t, group)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("updating a permissions group that doesn't exist should return an error", func(t *testing.T) {
		err := client.UpdatePermissionsGroup(1000, PermissionsGroupRequest{
			Name: acctest.RandString(10),
		})

		assert.ErrorContains(t, err, "Not found.")
	})

	t.Run("deleting a permissions group that doesn't exist should be gracefully handled", func(t *testing.T) {
		err := client.DeletePermissionsGroup(1000)

		assert.NoError(t, err)
	})
}
