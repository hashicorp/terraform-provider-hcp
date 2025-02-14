// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	provider "github.com/hashicorp/terraform-provider-hcp/internal/provider"
	providersdkv2 "github.com/hashicorp/terraform-provider-hcp/internal/providersdkv2"
	"github.com/hashicorp/terraform-provider-hcp/version"
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

	var serveOpts []tf6server.ServeOpt

	if debugMode {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve("registry.terraform.io/hashicorp/hcp", provider, serveOpts...)
	if err != nil {
		log.Fatal(err)
	}
}

func New() (func() tfprotov6.ProviderServer, error) {
	ctx := context.Background()

	// Upgrade the provider sdkv2 version to protocol 6
	upgradedSdkProvider, err := tf5to6server.UpgradeServer(
		context.Background(),
		providersdkv2.New()().GRPCProvider,
	)
	if err != nil {
		return nil, err
	}

	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer {
			return upgradedSdkProvider
		},
		providerserver.NewProtocol6(
			provider.NewFrameworkProvider(version.ProviderVersion)(),
		),
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		return nil, err
	}
	return muxServer.ProviderServer, nil
}
