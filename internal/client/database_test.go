package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient_GetDatabase(t *testing.T) {
	client, _ := NewClient(testServerUrl, testUsername, testPassword, testHeaders)

	t.Run("requesting a database that exists should return that database", func(t *testing.T) {
		database, err := client.GetDatabase(1)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), database.Id)
	})

	t.Run("requesting a database that doesn't exist should return an error", func(t *testing.T) {
		database, err := client.GetDatabase(1000)

		assert.Nil(t, database)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestClient_Database_H2(t *testing.T) {
	client, _ := NewClient(testServerUrl, testUsername, testPassword, testHeaders)

	t.Run("creating a database with the DB string missing should return an error", func(t *testing.T) {
		databaseId, err := client.CreateDatabase(DatabaseRequest{
			Engine:  EngineH2,
			Name:    "Test H2",
			Details: map[string]interface{}{},
		})

		assert.Zero(t, databaseId)
		assert.Error(t, err)
	})

	t.Run("creating a valid database", func(t *testing.T) {
		dbConnString := "zip:/app/metabase.jar!/sample-database.db;USER=GUEST;PASSWORD=guest"
		databaseId, err := client.CreateDatabase(DatabaseRequest{
			Engine: EngineH2,
			Name:   "Test H2",
			Details: map[string]interface{}{
				"db": &dbConnString,
			},
		})

		assert.NoError(t, err)
		assert.NotZero(t, databaseId)

		t.Run("you should be able to fetch the database", func(t *testing.T) {
			database, err := client.GetDatabase(databaseId)

			assert.NoError(t, err)
			assert.Equal(t, "h2", database.Engine)
			assert.Equal(t, dbConnString, ((*database.Details)["db"]).(string))
		})

		t.Run("you should be able to update the database", func(t *testing.T) {
			err := client.UpdateDatabase(databaseId, DatabaseRequest{
				Engine: EngineH2,
				Name:   "Test H2 (Updated)",
				Details: map[string]interface{}{
					"db": &dbConnString,
				},
			})

			assert.NoError(t, err)
		})

		t.Run("you should be able to delete the database", func(t *testing.T) {
			err := client.DeleteDatabase(databaseId)

			assert.NoError(t, err)
		})
	})
}
