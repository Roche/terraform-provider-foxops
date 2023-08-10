package main

import (
	"context"
	"flag"
	"log"

	"github.com/Roche/terraform-provider-foxops/internal/client"
	"github.com/Roche/terraform-provider-foxops/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	version = "dev"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// TODO: Update this string with the published name of your provider.
		Address: "registry.terraform.io/Roche/foxops",
		Debug:   debug,
	}

	err := providerserver.Serve(
		context.Background(),
		provider.New(
			provider.Version(version),
			provider.ClientConstructor(
				func(
					ce provider.ClientEndpoint,
					ct provider.ClientToken,
					v provider.Version,
				) provider.FoxopsClient {
					return client.New(ce, ct, v)
				},
			),
			[]func() datasource.DataSource{
				provider.NewIncarnationDataSource,
			},
			[]func() resource.Resource{
				provider.NewIncarnationResource,
			},
		),
		opts,
	)

	if err != nil {
		log.Fatal(err.Error())
	}
}
