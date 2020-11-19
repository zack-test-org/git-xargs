package main

import (
	"github.com/gruntwork-io/prototypes/tf-provider-hackday/secrets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main()  {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return secrets.Provider()
		},
	})
}

