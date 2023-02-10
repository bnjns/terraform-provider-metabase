package schema

import (
	dSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"terraform-provider-metabase/internal/validators"
)

func PermissionsGroupResource() rSchema.Schema {
	return rSchema.Schema{
		Description: "Allows for creating and managing permissions groups (user groups) in Metabase.",
		Attributes: map[string]rSchema.Attribute{
			"id": rSchema.Int64Attribute{
				Description: "The ID of the permissions group.",
				Computed:    true,
			},
			"name": rSchema.StringAttribute{
				Description: "The name of the permissions group.",
				Required:    true,
				Validators: []validator.String{
					validators.NotEmptyStringValidator(),
				},
			},
		},
	}
}

func PermissionsGroupDataSource() dSchema.Schema {
	return dSchema.Schema{
		Description: "Gets the details of the provided permissions (user) group.",
		Attributes: map[string]dSchema.Attribute{
			"id": dSchema.Int64Attribute{
				Description: "The ID of the permissions group.",
				Required:    true,
			},
			"name": dSchema.StringAttribute{
				Description: "The name of the permissions group.",
				Computed:    true,
			},
		},
	}
}
