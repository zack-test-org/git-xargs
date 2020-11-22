package secure_secrets

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"os"
	"strings"
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSecretValueCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// How to do logging in a Terraform provider... Note that you need to run with the environment variable
	// TF_LOG=debug to actually see this logs when you run plan or apply.
	log.Printf("[DEBUG] resourceSecretValueCreate called\n")

	client, ok := m.(*secretsmanager.SecretsManager)
	if !ok {
		return diag.Errorf("Didn't get expected SecretsManager client")
	}

	name, err := getRequiredString(d, "name")
	if err != nil {
		return diag.FromErr(err)
	}

	description, err := getOptionalString(d, "description")
	if err != nil {
		return diag.FromErr(err)
	}

	kmsKeyId, err := getOptionalString(d, "kms_key_id")
	if err != nil {
		return diag.FromErr(err)
	}

	value, err := getSecureSecretValueRequired(name)
	if err != nil {
		return diag.FromErr(err)
	}

	input := secretsmanager.CreateSecretInput{
		Name:         aws.String(name),
		Description:  description,
		KmsKeyId:     kmsKeyId,
		SecretString: aws.String(value),
	}

	secret, err := client.CreateSecret(&input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(aws.StringValue(secret.ARN))

	// You're supposed to read your resource after creating it so the state ends up the same way after creation and
	// read. Perhaps this minimizes spurious diffs?
	return resourceSecretValueRead(ctx, d, m)
}

func resourceSecretValueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client, ok := m.(*secretsmanager.SecretsManager)
	if !ok {
		return diag.Errorf("Didn't get expected SecretsManager client")
	}

	input := secretsmanager.DescribeSecretInput{
		SecretId: aws.String(d.Id()),
	}

	secret, err := client.DescribeSecret(&input)
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
	if err := d.Set("description", secret.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("kms_key_id", secret.KmsKeyId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("arn", secret.ARN); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSecretValueUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*secretsmanager.SecretsManager)
	if !ok {
		return diag.Errorf("Didn't get expected SecretsManager client")
	}

	hasChanges, err := secretHasChanges(client, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	if hasChanges {
		description, err := getOptionalString(d, "description")
		if err != nil {
			return diag.FromErr(err)
		}

		kmsKeyId, err := getOptionalString(d, "kms_key_id")
		if err != nil {
			return diag.FromErr(err)
		}

		input := secretsmanager.UpdateSecretInput{
			SecretId:    aws.String(d.Id()),
			Description: description,
			KmsKeyId:    kmsKeyId,
		}

		name, err := getRequiredString(d, "name")
		if err != nil {
			return diag.FromErr(err)
		}

		value := getSecureSecretValueOptional(name)

		if value != "" {
			input.SecretString = aws.String(value)
		}

		if _, err := client.UpdateSecret(&input); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceSecretValueRead(ctx, d, m)
}

func resourceSecretValueDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client, ok := m.(*secretsmanager.SecretsManager)
	if !ok {
		return diag.Errorf("Didn't get expected SecretsManager client")
	}

	input := secretsmanager.DeleteSecretInput{
		SecretId: aws.String(d.Id()),
		// TODO: should this be here? This is covenient for testing, but in prod, we may want to support a recovery period!
		ForceDeleteWithoutRecovery: aws.Bool(true),
	}

	if _, err := client.DeleteSecret(&input); err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func secretHasChanges(client *secretsmanager.SecretsManager, d *schema.ResourceData, m interface{}) (bool, error) {
	if d.HasChange("description") || d.HasChange("kms_key_id") {
		return true, nil
	}

	name, err := getRequiredString(d, "name")
	if err != nil {
		return false, err
	}

	value := getSecureSecretValueOptional(name)
	if value == "" {
		return false, nil
	}

	valueFromAws, err := getSecretValueFromAws(client, d)
	if err != nil {
		return false, err
	}

	return value != valueFromAws, nil
}

func getSecretValueFromAws(client *secretsmanager.SecretsManager, d *schema.ResourceData) (string, error) {
	input := secretsmanager.GetSecretValueInput{
		SecretId: aws.String(d.Id()),
	}
	secret, err := client.GetSecretValue(&input)
	if err != nil {
		return "", err
	}

	return aws.StringValue(secret.SecretString), nil
}

func getSecureSecretValueRequired(secretName string) (string, error) {
	value := getSecureSecretValueOptional(secretName)
	if value == "" {
		return "", fmt.Errorf("No value found for secret '%s'. You must set the value of this secret using the environment variable '%s'. Using environment variables ensures the secrets stay out of Terraform state.", secretName, secureSecretEnvVarName(secretName))
	}

	return value, nil
}

func getSecureSecretValueOptional(secretName string) string {
	envVarName := secureSecretEnvVarName(secretName)
	return os.Getenv(envVarName)
}

func secureSecretEnvVarName(secretName string) string {
	// Env vars cannot have dashes
	return fmt.Sprintf("SECRET_%s", strings.ReplaceAll(secretName, "-", "_"))
}
