package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshallJson(t *testing.T) {
	t.Parallel()

	t.Run("null string should return nil and no errors", func(t *testing.T) {
		configString := types.StringNull()

		config, err := UnmarshallJson(configString)

		assert.Nil(t, err)
		assert.Nil(t, config)
	})

	t.Run("invalid JSON should return error", func(t *testing.T) {
		configString := types.StringValue("invalid json")

		config, err := UnmarshallJson(configString)

		assert.Nil(t, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error processing database configuration")
	})

	t.Run("valid JSON should be unmarshalled", func(t *testing.T) {
		configString := types.StringValue(`
{
	"first": "value",
	"second": 2
}
`)

		config, err := UnmarshallJson(configString)

		assert.Nil(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, 2, len(config))
		assert.Equal(t, "value", config["first"])
		assert.Equal(t, float64(2), config["second"])
	})
}
