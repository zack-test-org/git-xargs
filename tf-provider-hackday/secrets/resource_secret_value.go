package secrets

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func resourceSecretValue() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretValueCreate,
		ReadContext:   resourceSecretValueRead,
		UpdateContext: resourceSecretValueUpdate,
		DeleteContext: resourceSecretValueDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				// If anyone  updates the name attribute, we delete the old secret and create a totally new one.
				ForceNew: true,
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceSecretValueCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// How to do logging in a Terraform provider... Note that you need to run with the environment variable
	// TF_LOG=debug to actually see this logs when you run plan or apply.
	log.Printf("[DEBUG] resourceSecretValueCreate called\n")

	client, ok := m.(*MockClient)
	if !ok {
		return diag.Errorf("Didn't get expected MockClient")
	}

	name, err := getRequiredString(d, "name")
	if err != nil {
		return diag.FromErr(err)
	}

	value, err := getRequiredString(d, "value")
	if err != nil {
		return diag.FromErr(err)
	}

	secret, err := client.CreateSecret(name, value)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(secret.Id)

	// You're supposed to read your resource after creating it so the state ends up the same way after creation and
	// read. Perhaps this minimizes spurious diffs?
	return resourceSecretValueRead(ctx, d, m)
}

func resourceSecretValueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client, ok := m.(*MockClient)
	if !ok {
		return diag.Errorf("Didn't get expected MockClient")
	}

	secret, err := client.GetSecretById(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// I tried to find a way to avoid having to manually set each and every field, such as passing in the secret struct
	// directly, using nested values in the schema, using mapstructure to convert to a map, and so on. None of them
	// work well. I browsed the code for several other providers, and all of them end up setting all this data manually.
	// It's verbose and not at all DRY... But I'm not sure there's a cleaner way.
	if err := d.Set("name", secret.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("value", secret.Value); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSecretValueUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*MockClient)
	if !ok {
		return diag.Errorf("Didn't get expected MockClient")
	}

	if d.HasChange("value") {
		name, err := getRequiredString(d, "name")
		if err != nil {
			return diag.FromErr(err)
		}

		value, err := getRequiredString(d, "value")
		if err != nil {
			return diag.FromErr(err)
		}

		if _, err := client.UpdateSecret(name, value); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceSecretValueRead(ctx, d, m)
}

func resourceSecretValueDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client, ok := m.(*MockClient)
	if !ok {
		return diag.Errorf("Didn't get expected MockClient")
	}

	if err := client.DeleteSecretById(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
