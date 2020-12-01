package main

import (
	secure_secrets "github.com/gruntwork-io/prototypes/secure-secrets-provider/secure-secrets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main()  {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return secure_secrets.Provider()
		},
	})
}

