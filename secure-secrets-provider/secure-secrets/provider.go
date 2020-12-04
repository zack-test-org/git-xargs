package secure_secrets

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"securesecrets_value": resourceSecretValue(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	region, ok := d.Get("region").(string)
	if !ok {
		return nil, diag.Errorf("The 'region' param in the securesecrets provider must be a string")
	}

	sess, err := NewAuthenticatedSessionFromDefaultCredentials(region)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return secretsmanager.New(sess), nil
}

// NewAuthenticatedSessionFromDefaultCredentials gets an AWS Session, checking that the user has credentials properly configured in their environment.
func NewAuthenticatedSessionFromDefaultCredentials(region string) (*session.Session, error) {
	sess, err := session.NewSession(aws.NewConfig().WithRegion(region))
	if err != nil {
		return nil, err
	}

	if _, err = sess.Config.Credentials.Get(); err != nil {
		return nil, err
	}

	return sess, nil
}