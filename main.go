package main

import (
	// "terraform-provider-zosmf/zosmf"

	"terraform-provider-zosmf/zosmf"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return zosmf.Provider()
		},
	})
	// plugin.Serve(&plugin.ServeOpts{
	// 	ProviderFunc: func() *schema.Provider {
	// 		return hashicups.Provider()
	// 	},
	// })

	// var debugMode bool

	// flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	// flag.Parse()

	// opts := &plugin.ServeOpts{
	// 	ProviderFunc: func() *schema.Provider {
	// 		return zosmf.Provider()
	// 	},
	// }

	// if debugMode {
	// 	// TODO: update this string with the full name of your provider as used in your configs
	// 	err := plugin.Debug(context.Background(), "registry.terraform.io/my-org/my-provider", opts)
	// 	if err != nil {
	// 		log.Fatal(err.Error())
	// 	}
	// 	return
	// }

	// plugin.Serve(opts)
}
