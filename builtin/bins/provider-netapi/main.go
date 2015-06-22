package main

import (
	"github.com/hashicorp/terraform/builtin/providers/netapi"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: netapi.Provider,
	})
}
