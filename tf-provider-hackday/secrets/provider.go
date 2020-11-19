package secrets

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("FOO_USERNAME", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("FOO_PASSWORD", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"secrets_value": resourceSecretValue(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"secrets_value": dataSourceSecretValue(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	username := d.Get("username")
	password := d.Get("password")

	usernameStr, ok := username.(string)
	if !ok {
		return nil, diag.Errorf("The username param in the secrets provider must be a string")
	}

	passwordStr, ok := password.(string)
	if !ok {
		return nil, diag.Errorf("The password param in the secrets provider must be a string")
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client, err := NewClient(usernameStr, passwordStr)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create secrets client",
			Detail:   "Unable to auth to MockClient",
		})
	}

	return client, diags
}
