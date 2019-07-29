package main

import (
	"github.com/hive-io/terraform-provider-hiveio/hiveio"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return hiveio.Provider()
		},
	})
}
