package schema

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	rSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-metabase/internal/validators"
)

var DatabaseScheduleType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"day":    types.StringType,
		"frame":  types.StringType,
		"hour":   types.Int64Type,
		"minute": types.Int64Type,
		"type":   types.StringType,
	},
}

func DatabaseResource() rSchema.Schema {
	return rSchema.Schema{
		Description: "Allows you to manage a database configuration",
		Attributes: map[string]rSchema.Attribute{
			"id": rSchema.Int64Attribute{
				Description: "The ID of the database.",
				Computed:    true,
			},
			"engine": rSchema.StringAttribute{
				Description: "The engine type of the database.",
				Required:    true,
				Validators: []validator.String{
					validators.IsKnownDatabaseEngineValidator(),
				},
			},
			"name": rSchema.StringAttribute{
				Description: "The name of the database.",
				Required:    true,
			},
			"features": rSchema.ListAttribute{
				ElementType: types.StringType,
				Description: "The features this database engine supports.",
				Computed:    true,
			},
			"details": rSchema.StringAttribute{
				Description:         "Serialised JSON string containing the configuration options for the database engine. Use details_secure for any sensitive configuration details (eg, password).",
				MarkdownDescription: "Serialised JSON string containing the configuration options for the database engine. Use `details_secure` for any sensitive configuration details (eg, password).",
				Optional:            true,
			},
			"details_secure": rSchema.StringAttribute{
				Description: "Serialised JSON string containing any sensitive configuration options for the database engine.",
				Optional:    true,
				Sensitive:   true,
			},
			"schedules": rSchema.MapAttribute{
				ElementType: DatabaseScheduleType,
				Description: "The schedules used to sync the database.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}
