package utils

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func UnmarshallJson(config types.String) (map[string]interface{}, error) {
	if config.IsNull() {
		return nil, nil
	} else {
		configUnmarshalled := make(map[string]interface{})
		err := json.Unmarshal([]byte(config.ValueString()), &configUnmarshalled)
		if err != nil {
			return nil, fmt.Errorf("error processing database configuration: %e", err)
		}

		return configUnmarshalled, nil
	}
}
