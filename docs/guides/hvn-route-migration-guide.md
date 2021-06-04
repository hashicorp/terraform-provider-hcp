---
subcategory: ""
page_title: "HVN Route Migration Guide - HCP Provider"
description: |-
    An guide to migrating HCP networking resources to use HVN routes.
---

# Introducing HVN routes

The HVN route is a new resource that belongs to an HVN. It contains a CIDR block and targets a networking connection: 
either a peering or transit gateway attachment.

HVN routes provide a general view on how an HVN's traffic is routed across all networking connections and create a flexible way of managing these routing rules.

## Migrating existing peerings and transit gateway attachments

There are two ways to migrate existing peerings and transit gateway attachments managed by Terraform:

 1. Recreate Resources with Updated Schema
    * This option is quicker but will result in downtime and possible data loss. Best for test environments. Will allow you to specify human-readable ids for the resources.
    * Comment out all `hcp_aws_network_peering` and `hcp_aws_transit_gateway_attachment` resources.
    * Run `terraform apply` to destroy currently existing connections.
    * Uncomment and update all `hcp_aws_network_peering` and `hcp_aws_transit_gateway_attachment` resource definitions to match the new schema. 
    * Add corresponding `hcp_hvn_route` resources for each CIDR targeting corresponding peering connections or transit gateway attachment.
    * Run `terraform apply` to recreate connections.

 2. Re-Import with Updated Syntax:
    * This option allows you to avoid downtime or data loss.
    * Update any `hcp_aws_network_peering` and `hcp_aws_transit_gateway_attachment` resource definitions to match the new schema. All values needed can be found on the details pages of Peerings and TGW attachment in the HCP Portal.
    * Add corresponding `hcp_hvn_route` resources for each CIDR targeting corresponding peering connections or transit gateway attachments.
    * Run `terraform import hcp_hvn_route.<route-name> <hvn-id>:<hvn-route-id>` for each `hcp_hvn_route`. The `<hvn-route-id>` can be found on the details pages of the corresponding HVN connection in the HCP Portal.
    * Run `terraform plan` and make sure that there are no changes detected by the Terraform.

The examples below walk through the schema upgrade and re-import steps.

### Peering example

Given:
```terraform
resource "hcp_hvn" "hvn" {
  hvn_id         = "prod-hvn"
  region         = "us-west-2"
  cloud_provider = "aws"
}

resource "hcp_aws_network_peering" "peering" {
  hvn_id              = hcp_hvn.hvn.hvn_id
  peer_vpc_id         = "vpc-845f29fc"
  peer_account_id     = "572816266891"
  peer_vpc_region     = "us-west-2"
  peer_vpc_cidr_block = "172.31.0.0/16"
}
```

Rewrite it to the new schema and add corresponding HVN route:
```terraform
resource "hcp_hvn" "hvn" {
  hvn_id         = "prod-hvn"
  region         = "us-west-2"
  cloud_provider = "aws"
}

resource "hcp_aws_network_peering" "peering" {
  hvn_id = hcp_hvn.hvn.hvn_id
  // add `peering_id` that you can find in the HCP Portal
  peering_id      = "f03324a9-4377-4a54-9c15-958fd07ad77b"
  peer_vpc_id     = "vpc-845f29fc"
  peer_account_id = "572816266891"
  peer_vpc_region = "us-west-2"
  // remove `peer_vpc_cidr_block`
  // peer_vpc_cidr_block = "172.31.0.0/16"
}

// Add a `hcp_hvn_route` resource for the peering's CIDR
resource "hcp_hvn_route" "peering-route" {
  hvn_link = hcp_hvn.hvn.self_link
  // you can find this ID in the HCP Portal in the peering details page in the list of routes
  hvn_route_id     = "a8dda9a8-0f69-4fa0-b38c-55be302fdddb"
  destination_cidr = "172.31.0.0/16"
  target_link      = hcp_aws_network_peering.peering.self_link
}
```

Run `import` for the `hcp_hvn_route`:
```shell
$ terraform import hcp_hvn_route.peering-route prod-hvn:a8dda9a8-0f69-4fa0-b38c-55be302fdddb
```

Run `terraform plan` to make sure there are no changes detected by the Terraform:
```shell
$ terraform plan
No changes. Infrastructure is up-to-date.
```

### Transit gateway attachment example

Given:
```terraform
resource "hcp_hvn" "hvn" {
  hvn_id         = "prod-hvn"
  region         = "us-west-2"
  cloud_provider = "aws"
}

resource "hcp_aws_transit_gateway_attachment" "prod" {
  hvn_id                        = hcp_hvn.hvn.hvn_id
  transit_gateway_attachment_id = "prod-tgw-attachment"
  transit_gateway_id            = "tgw-0ee94b1a1167cf89d"
  resource_share_arn            = "arn:aws:ram:us-west-2:..."
  destination_cidrs             = ["10.1.0.0/24", "10.2.0.0/24"]
}
```

Rewrite it to the new schema and add corresponding HVN route:
```terraform
resource "hcp_hvn" "hvn" {
  hvn_id         = "prod-hvn"
  region         = "us-west-2"
  cloud_provider = "aws"
}

resource "hcp_aws_transit_gateway_attachment" "prod" {
  hvn_id                        = hcp_hvn.hvn.hvn_id
  transit_gateway_attachment_id = "prod-tgw-attachment"
  transit_gateway_id            = "tgw-0ee94b1a1167cf89d"
  resource_share_arn            = "arn:aws:ram:us-west-2:..."
  // remove `destination_cidrs`
  // destination_cidrs             = ["10.1.0.0/24", "10.2.0.0/24"]
}

// add a new `hcp_hvn_route` for each CIDR associated with the transit gateway attachment
resource "hcp_hvn_route" "tgw-route-1" {
  hvn_link = hcp_hvn.hvn.self_link
  // you can find this ID in the HCP Portal in the TGW attachment details page in the list of Routes
  hvn_route_id     = "35392425-215a-44ec-bbd0-051bb777ce5f"
  destination_cidr = "10.1.0.0/24"
  target_link      = hcp_aws_transit_gateway_attachment.prod.self_link
}

resource "hcp_hvn_route" "tgw-route-2" {
  hvn_link = hcp_hvn.hvn.self_link
  // you can find this ID in the HCP Portal in the transit gateway attachment details page in the list of routes
  hvn_route_id     = "9867959a-d81b-4e52-ae8e-ca56f9dd06fc"
  destination_cidr = "10.2.0.0/24"
  target_link      = hcp_aws_transit_gateway_attachment.prod.self_link
}
```

Run `import` for each `hcp_hvn_route` you've added:
```shell
$ terraform import hcp_hvn_route.tgw-route-1 prod-hvn:35392425-215a-44ec-bbd0-051bb777ce5f
...

$ terraform import hcp_hvn_route.tgw-route-2 prod-hvn:9867959a-d81b-4e52-ae8e-ca56f9dd06fc
...
```

Run `terraform plan` to make sure there are no changes detected by the Terraform:
```shell
$ terraform plan
No changes. Infrastructure is up-to-date.
```