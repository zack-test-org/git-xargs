package secrets

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func dataSourceSecretValue() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretValueRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceSecretValueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// How to do logging in a Terraform provider... Note that you need to run with the environment variable
	// TF_LOG=debug to actually see this logs when you run plan or apply.
	log.Printf("[DEBUG] dataSourceSecretValueRead called\n")

	client, ok := m.(*MockClient)
	if !ok {
		return diag.Errorf("Didn't get expected MockClient")
	}

	var diags diag.Diagnostics

	name, err := getRequiredString(d, "name")
	if err != nil {
		return diag.FromErr(err)
	}

	secret, err := client.GetSecretByName(name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("value", secret.Value); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(secret.Id)

	return diags
}
