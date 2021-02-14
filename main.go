package main

import (
	"github.com/ggsood/terraform-provider-elasticsearch/v7/es"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: es.Provider,
	})
}
