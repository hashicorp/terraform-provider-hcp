---
page_title: "hcp_aws_transit_gateway_attachment Resource - terraform-provider-hcp"
subcategory: ""
description: |-
  The AWS Transit gateway attachment resource allows you to manage a Transit gateway attachment that attaches an HVN to a Transit gateway in AWS.
---

# Resource `hcp_aws_transit_gateway_attachment`

The AWS Transit gateway attachment resource allows you to manage a Transit gateway attachment that attaches an HVN to a Transit gateway in AWS.

## Example Usage

```terraform
provider "aws" {
  region = "us-west-2"
}

resource "hcp_hvn" "main" {
  hvn_id         = "main-hvn"
  cloud_provider = "aws"
  region         = "us-west-2"
  cidr_block     = "172.25.16.0/20"
}

resource "aws_vpc" "example" {
  cidr_block = "172.31.0.0/16"
}

resource "aws_ec2_transit_gateway" "example" {
  tags = {
    Name = "example-tgw"
  }
}

resource "aws_ram_resource_share" "example" {
  name                      = "example-resource-share"
  allow_external_principals = true
}

resource "aws_ram_principal_association" "example" {
  resource_share_arn = aws_ram_resource_share.example.arn
  principal          = hcp_hvn.main.provider_account_id
}

resource "aws_ram_resource_association" "example" {
  resource_share_arn = aws_ram_resource_share.example.arn
  resource_arn       = aws_ec2_transit_gateway.example.arn
}

resource "hcp_aws_transit_gateway_attachment" "example" {
  depends_on = [
    aws_ram_principal_association.example,
    aws_ram_resource_association.example,
  ]

  hvn_id                        = hcp_hvn.main.hvn_id
  transit_gateway_attachment_id = "example-tgw-attachment"
  transit_gateway_id            = aws_ec2_transit_gateway.example.id
  resource_share_arn            = aws_ram_resource_share.example.arn
  destination_cidrs             = [aws_vpc.example.cidr_block]
}

resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "example" {
  transit_gateway_attachment_id = hcp_aws_transit_gateway_attachment.example.provider_transit_gateway_attachment_id
}
```

## Schema

### Required

- **destination_cidrs** (List of String) The list of associated CIDR ranges. Traffic from these CIDRs will be allowed for all resources in the HVN. Traffic to these CIDRs will be routed into this Transit gateway attachment.
- **hvn_id** (String) The ID of the HashiCorp Virtual Network (HVN).
- **resource_share_arn** (String, Sensitive) The Amazon Resource Name (ARN) of the Resource Share that is needed to grant HCP access to the Transit gateway in AWS. The Resource Share should be associated with the HCP AWS account principal (see [aws_ram_principal_association](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ram_principal_association)) and the Transit gateway resource (see [aws_ram_resource_association](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ram_resource_association))
- **transit_gateway_attachment_id** (String) The user-settable name of the Transit gateway attachment in HCP.
- **transit_gateway_id** (String) The ID of the Transit gateway in AWS.

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **created_at** (String) The time that the Transit gateway attachment was created.
- **expires_at** (String) The time after which the Transit gateway attachment will be considered expired if it hasn't transitioned into 'Accepted' or 'Active' state.
- **organization_id** (String) The ID of the HCP organization where the Transit gateway attachment is located. Always matches the HVN's organization.
- **project_id** (String) The ID of the HCP project where the Transit gateway attachment is located. Always matches the HVN's project.
- **provider_transit_gateway_attachment_id** (String) The Transit gateway attachment ID used by AWS.
- **state** (String) The state of the Transit gateway attachment.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **default** (String)
- **delete** (String)


