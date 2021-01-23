package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hive-io/terraform-provider-hiveio/hiveio"
)

//go:generate terraform fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return hiveio.Provider()
		},
	})
}
