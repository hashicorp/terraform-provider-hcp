// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-provider-hcp/internal/provider"
	"github.com/hashicorp/terraform-provider-hcp/version"
	"github.com/hashicorp/vault/sdk/plugin"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	provider, err := New()
	if err != nil {
		return
	}

	if debugMode {
		opts := &plugin.ServeOpts{
			Debug: true,
		}
		tf5server.Serve("registry.terraform.io/hashicorp/hcp", provider, opts)
		return
	}

	tf5server.Serve("registry.terraform.io/hashicorp/hcp", provider)
}

// TODO:
// - Add user agent back in
// - Add validators
func New() (func() tfprotov5.ProviderServer, error) {
	ctx := context.Background()
	providers := []func() tfprotov5.ProviderServer{
		provider.New()().GRPCProvider,
		providerserver.NewProtocol5(
			provider.NewFrameworkProvider(version.ProviderVersion)(),
		),
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		return nil, err
	}
	return muxServer.ProviderServer, nil
}
