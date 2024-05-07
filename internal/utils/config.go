package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
)

func GetConfigValue(cfg types.String, envName string) string {
	if cfg.IsNull() || cfg.IsUnknown() {
		return os.Getenv(envName)
	} else {
		return cfg.ValueString()
	}
}
