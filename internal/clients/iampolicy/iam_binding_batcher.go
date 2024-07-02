// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iampolicy

import (
	"context"
	"log"
	"sync"
	"time"

	iamModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
	"github.com/hashicorp/terraform-provider-hcp/internal/customdiags"
	"golang.org/x/exp/maps"
)

const (
	// policyBatchDuration is the duration we batch getting or setting an IAM Policy
	policyBatchDuration = 1 * time.Second
)

func init() {
	// Create the singleton on package initialization.
	bindingsBatcher = newIamBindingsBatcher()
}

// bindingsBatcher is the singleton iamBindingsBatcher
var bindingsBatcher *iamBindingsBatcher

// iamBindingsBatcher allows batching changes to a given resource's IAM policy.
type iamBindingsBatcher struct {
	batches map[string]*iamBindingBatcher
	sync.Mutex
}

// newIamBindingsBatcher creates a new newIamBindingsBatcher.
func newIamBindingsBatcher() *iamBindingsBatcher {
	return &iamBindingsBatcher{
		batches: make(map[string]*iamBindingBatcher, 16),
	}
}

// getBatch retrieves an iamBindingBatcher for the given resource. It takes the
// ResourceIamUpdater for that resource.
func (b *iamBindingsBatcher) getBatch(updater ResourceIamUpdater) *iamBindingBatcher {
	b.Lock()
	defer b.Unlock()

	// Get an existing batcher
	key := updater.GetMutexKey()
	batch, ok := b.batches[key]
	if ok {
		return batch
	}

	// Create a new batcher
	batch = newIamBindingBatcher(updater)
	b.batches[key] = batch
	return batch

}

// iamBindingBatcher is used to batch changes to a resource's IAM policy.
type iamBindingBatcher struct {
	updater ResourceIamUpdater
	sync.Mutex

	getFuture *policyFuture
	setFuture *policyFuture
}

// newIamBindingBatcher takes the ResourceIamUpdater for the given resource and
// returns a batcher.
func newIamBindingBatcher(updater ResourceIamUpdater) *iamBindingBatcher {
	return &iamBindingBatcher{
		updater: updater,
	}
}

// GetPolicy retrieves the policy for the given resource. Multiple concurrent
// callers will be combined into a single request.
func (b *iamBindingBatcher) GetPolicy(ctx context.Context) *policyFuture {
	b.Lock()
	defer b.Unlock()

	// We have an existing future. Check if it is done.
	if b.getFuture != nil {
		select {
		case <-b.getFuture.doneCh:
		default:
			// It is not done so attach this request to the existing future
			return b.getFuture
		}
	}

	// This is either the first request or the existing future has already
	// completed.
	b.getFuture = newPolicyFuture()
	time.AfterFunc(policyBatchDuration, func() {
		b.getFuture.set(b.updater.GetResourceIamPolicy(ctx))
	})

	return b.getFuture
}

// ModifyPolicy modifies the resource's IAM policy. Modifications can be made by
// passing either a set binding or a remove binding. Both expect a role and a
// single member. Multiple callers will be batched into a single update request.
func (b *iamBindingBatcher) ModifyPolicy(ctx context.Context, client *clients.Client,
	setBinding, removeBinding *models.HashicorpCloudResourcemanagerPolicyBinding) *policyFuture {

	b.Lock()
	defer b.Unlock()

	// We have an existing future. Check if it is done.
	if b.setFuture != nil {
		select {
		case <-b.setFuture.doneCh:
		default:
			// It is not done so attach this request to the existing future
			b.setFuture.addBindingModfiers(setBinding, removeBinding)
			return b.setFuture
		}
	}

	// This is either the first request or the existing future has already
	// completed.
	b.setFuture = newPolicyFuture()
	b.setFuture.addBindingModfiers(setBinding, removeBinding)
	time.AfterFunc(policyBatchDuration, func() {
		b.setFuture.executeModifers(ctx, b.updater, client)
	})

	return b.setFuture
}

// policyFuture is a future for interacting with a resource's IAM policy.
type policyFuture struct {
	p      *models.HashicorpCloudResourcemanagerPolicy
	d      diag.Diagnostics
	doneCh chan struct{}

	// Store the modifiers
	setters  []*models.HashicorpCloudResourcemanagerPolicyBinding
	removers []*models.HashicorpCloudResourcemanagerPolicyBinding
}

func newPolicyFuture() *policyFuture {
	return &policyFuture{
		doneCh: make(chan struct{}),
	}
}

// Get retrieves the resource's IAM Policy or returns an error that occurred.
// This is a blocking call.
func (f *policyFuture) Get() (p *models.HashicorpCloudResourcemanagerPolicy, d diag.Diagnostics) {
	<-f.doneCh
	return f.p, f.d
}

// set sets the results and unblocks any waiting callers on Get.
func (f *policyFuture) set(p *models.HashicorpCloudResourcemanagerPolicy, d diag.Diagnostics) {
	f.p = p
	f.d = d
	close(f.doneCh)
}

// addBindingModfiers adds binding modifiers to the future. All of them will be
// executed in a single batch.
func (f *policyFuture) addBindingModfiers(set, remove *models.HashicorpCloudResourcemanagerPolicyBinding) {
	if set != nil {
		f.setters = append(f.setters, set)
	}
	if remove != nil {
		f.removers = append(f.removers, remove)
	}
}

// executeModifers applies all modifiers that are set on the future.
func (f *policyFuture) executeModifers(ctx context.Context, u ResourceIamUpdater, client *clients.Client) {
	if diags := f.validateModifiers(); diags.HasError() {
		f.set(nil, diags)
		return
	}

	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		// Get the existing policy
		ep, diags := u.GetResourceIamPolicy(ctx)
		if diags.HasError() {
			f.set(nil, diags)
			return
		}

		// Determine the principal's we need to lookup their type.
		principalSet, diags := f.getPrincipals(ctx, client)
		if diags.HasError() {
			f.set(nil, diags)
			return
		}

		// Remove any bindings needed
		bindings := ToMap(ep)
		for _, rm := range f.removers {
			if members, ok := bindings[rm.RoleID]; ok {
				delete(members, rm.Members[0].MemberID)
				if len(members) == 0 {
					delete(bindings, rm.RoleID)
				}
			}
		}

		// Go through the setters and apply them
		for _, s := range f.setters {
			members, ok := bindings[s.RoleID]
			if !ok {
				members = make(map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType, 4)
				bindings[s.RoleID] = members
			}

			members[s.Members[0].MemberID] = principalSet[s.Members[0].MemberID]
		}

		// Apply the policy
		ep, diags = u.SetResourceIamPolicy(ctx, FromMap(ep.Etag, bindings))
		if diags.HasError() {
			if customdiags.HasConflictError(diags) {
				// Policy object has changed since it was last gotten and the etag is now different.
				// Continuously retry getting and setting the policy with an increasing backoff period until the maximum backoff period is reached.
				if backoff > maxBackoff {
					log.Printf("[DEBUG]: Maximum backoff time reached. Aborting operation.")
					f.set(nil, diags)
					return
				}
				log.Printf("[DEBUG]: Operation failed due to conflicts. Operation will be restarted after %s", backoff)
				// Pause the execution for the duration specified by the current backoff time.
				time.Sleep(backoff)
				// Double the backoff time to increase the delay for the next retry.
				backoff *= 2
				continue
			}
			f.set(nil, diags)
			return
		}

		// Successfully applied policy
		f.set(ep, nil)
		return
	}
}

// getPrincipals returns a map of principal_id to binding type. The binding type
// is determined by doing a batch lookup of the principal's by ID.
func (f *policyFuture) getPrincipals(ctx context.Context, client *clients.Client) (map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType, diag.Diagnostics) {
	var diags diag.Diagnostics
	principalSet := make(map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType, 32)
	for _, set := range f.setters {
		principalSet[set.Members[0].MemberID] = nil
	}

	principals, err := clients.BatchGetPrincipals(ctx, client, maps.Keys(principalSet), iamModels.HashicorpCloudIamPrincipalViewPRINCIPALVIEWBASIC.Pointer())
	if err != nil {
		diags.AddError("failed to batch get principals", err.Error())
		return nil, diags
	}

	for _, p := range principals {
		t, err := clients.IamPrincipalTypeToBindingType(p)
		if err != nil {
			diags.AddError("Error converting principal types", err.Error())
			return nil, diags
		}

		principalSet[p.ID] = t
	}

	return principalSet, diags
}

// validateModifiers validates that all the modifier functions passed are valid.
func (f *policyFuture) validateModifiers() diag.Diagnostics {
	var diags diag.Diagnostics
	for _, rm := range f.removers {
		if rm.RoleID == "" || len(rm.Members) != 1 || rm.Members[0].MemberID == "" {
			diags.Append(diag.NewErrorDiagnostic("invalid binding remover", "either has blank role, or invalid members"))
			return diags
		}
	}

	for _, set := range f.setters {
		if set.RoleID == "" || len(set.Members) != 1 || set.Members[0].MemberID == "" {
			diags.Append(diag.NewErrorDiagnostic("invalid binding setter", "either has blank role, or invalid members"))
			return diags
		}
	}

	return diags
}
