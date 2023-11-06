// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// iampolicy is a helper package for creating Terraform resources that modify an
// HCP Resource's IAM Policy. By implementing a single interface, a resource can
// provide an authoratative Policy resource and a Binding resource. Using this
// package simplifies implementation, provides a performance optimized
// experience, and providers a consistent interface to the provider users.
//
// For examples of how to implement the interface, see
// hcp_project_iam_policy/binding.
package iampolicy
