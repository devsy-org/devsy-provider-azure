package azure

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/google/uuid"
)

// virtualMachineContributorRoleID is the well-known built-in role
// "Virtual Machine Contributor" — grants Microsoft.Compute/virtualMachines/*
// including deallocate, which lets the VM stop itself via managed identity.
const virtualMachineContributorRoleID = "9980e02c-c2be-4d73-94e8-173b1dc7cf3c"

// assignVMSelfContributorRole grants the VM's system-assigned managed identity
// the Virtual Machine Contributor role scoped to itself, so the in-VM
// inactivity timer can call `${AZURE_PROVIDER} stop` and deallocate.
func assignVMSelfContributorRole(ctx context.Context, azureProvider *AzureProvider) error {
	vm, err := getVirtualMachine(ctx, azureProvider)
	if err != nil {
		return fmt.Errorf("get virtual machine: %w", err)
	}
	if err := validateVMIdentity(vm); err != nil {
		return err
	}

	rolesClient, err := armauthorization.NewRoleAssignmentsClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return err
	}

	return createSelfRoleAssignment(ctx, rolesClient, vm, azureProvider.Config.SubscriptionID)
}

func validateVMIdentity(vm *armcompute.VirtualMachine) error {
	if vm.Identity == nil || vm.Identity.PrincipalID == nil {
		return fmt.Errorf("virtual machine has no system-assigned identity")
	}
	if vm.ID == nil {
		return fmt.Errorf("virtual machine has no resource ID")
	}
	return nil
}

func createSelfRoleAssignment(
	ctx context.Context,
	client *armauthorization.RoleAssignmentsClient,
	vm *armcompute.VirtualMachine,
	subscriptionID string,
) error {
	roleDefinitionID := fmt.Sprintf(
		"/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s",
		subscriptionID,
		virtualMachineContributorRoleID,
	)

	_, err := client.Create(
		ctx,
		*vm.ID,
		uuid.NewString(),
		armauthorization.RoleAssignmentCreateParameters{
			Properties: &armauthorization.RoleAssignmentProperties{
				PrincipalID:      vm.Identity.PrincipalID,
				RoleDefinitionID: to.Ptr(roleDefinitionID),
				PrincipalType:    to.Ptr(armauthorization.PrincipalTypeServicePrincipal),
			},
		},
		nil,
	)
	if err == nil {
		return nil
	}

	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) && respErr.StatusCode == http.StatusConflict {
		return nil
	}
	return err
}

func getVirtualMachine(
	ctx context.Context,
	azureProvider *AzureProvider,
) (*armcompute.VirtualMachine, error) {
	vmClient, err := armcompute.NewVirtualMachinesClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := vmClient.Get(
		ctx,
		azureProvider.Config.ResourceGroup,
		azureProvider.Config.MachineID,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.VirtualMachine, nil
}
