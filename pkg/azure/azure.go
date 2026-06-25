package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/devsy-org/devsy-provider-azure/pkg/options"
	"github.com/devsy-org/devsy/pkg/client"
)

type AzureProvider struct {
	Config           *options.Options
	Cred             *azidentity.DefaultAzureCredential
	WorkingDirectory string
}

func NewProvider() (*AzureProvider, error) {
	config, err := options.FromEnv(false)
	if err != nil {
		return nil, err
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("authentication failure: %w", err)
	}

	return &AzureProvider{
		Config: config,
		Cred:   cred,
	}, nil
}

func Create(ctx context.Context, azureProvider *AzureProvider) error {
	if _, err := createVirtualNetwork(ctx, azureProvider); err != nil {
		return fmt.Errorf("create virtual network: %w", err)
	}

	subnet, err := createSubnets(ctx, azureProvider)
	if err != nil {
		return fmt.Errorf("create subnet: %w", err)
	}

	publicIP, err := createPublicIP(ctx, azureProvider)
	if err != nil {
		return fmt.Errorf("create public IP address: %w", err)
	}

	nsg, err := createNetworkSecurityGroup(ctx, azureProvider)
	if err != nil {
		return fmt.Errorf("create network security group: %w", err)
	}

	netWorkInterface, err := createNetWorkInterface(
		ctx,
		azureProvider,
		networkInterfaceParams{
			subnetID:               *subnet.ID,
			publicIPID:             *publicIP.ID,
			networkSecurityGroupID: *nsg.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("create network interface: %w", err)
	}

	if _, err := createVirtualMachine(ctx, azureProvider, *netWorkInterface.ID); err != nil {
		return fmt.Errorf("create virtual machine: %w", err)
	}

	if err := assignVMSelfContributorRole(ctx, azureProvider); err != nil {
		return fmt.Errorf("assign self-stop role to VM identity: %w", err)
	}

	return nil
}

func Delete(ctx context.Context, azureProvider *AzureProvider) error {
	if err := deleteVirtualMachine(ctx, azureProvider); err != nil {
		return fmt.Errorf("delete virtual machine: %w", err)
	}

	if err := deleteDisk(ctx, azureProvider); err != nil {
		return fmt.Errorf("delete disk: %w", err)
	}

	if err := deleteNetWorkInterface(ctx, azureProvider); err != nil {
		return fmt.Errorf("delete network interface: %w", err)
	}

	if err := deleteNetworkSecurityGroup(ctx, azureProvider); err != nil {
		return fmt.Errorf("delete network security group: %w", err)
	}

	if err := deletePublicIP(ctx, azureProvider); err != nil {
		return fmt.Errorf("delete public IP address: %w", err)
	}

	if err := deleteSubnets(ctx, azureProvider); err != nil {
		return fmt.Errorf("delete subnet: %w", err)
	}

	if err := deleteVirtualNetWork(ctx, azureProvider); err != nil {
		return fmt.Errorf("delete virtual network: %w", err)
	}

	return nil
}

func Status(ctx context.Context, azureProvider *AzureProvider) (client.Status, error) {
	if !checkVirtualMachine(ctx, azureProvider) {
		return client.StatusNotFound, nil
	}

	vmClient, err := armcompute.NewVirtualMachinesClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return client.StatusNotFound, nil
	}

	resource, err := vmClient.InstanceView(
		ctx,
		azureProvider.Config.ResourceGroup,
		azureProvider.Config.MachineID,
		nil,
	)
	if err != nil {
		return client.StatusNotFound, nil
	}

	status := to.Ptr[string]("")
	if len(resource.Statuses) > 1 {
		status = resource.Statuses[1].DisplayStatus
	}

	switch *status {
	case "VM running":
		return client.StatusRunning, nil
	case "VM deallocated", "VM stopped":
		return client.StatusStopped, nil
	default:
		return client.StatusBusy, nil
	}
}

func Stop(ctx context.Context, azureProvider *AzureProvider) error {
	if !checkVirtualMachine(ctx, azureProvider) {
		return nil
	}

	vmClient, err := armcompute.NewVirtualMachinesClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return err
	}

	pollerResponse, err := vmClient.BeginDeallocate(
		ctx,
		azureProvider.Config.ResourceGroup,
		azureProvider.Config.MachineID,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	return err
}

func Start(ctx context.Context, azureProvider *AzureProvider) error {
	if !checkVirtualMachine(ctx, azureProvider) {
		return nil
	}

	vmClient, err := armcompute.NewVirtualMachinesClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return err
	}

	pollerResponse, err := vmClient.BeginStart(
		ctx,
		azureProvider.Config.ResourceGroup,
		azureProvider.Config.MachineID,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = pollerResponse.PollUntilDone(ctx, nil)
	return err
}

func GetInstanceIP(ctx context.Context, azureProvider *AzureProvider) (string, error) {
	publicIPAddressClient, err := armnetwork.NewPublicIPAddressesClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return "", err
	}

	resource, err := publicIPAddressClient.Get(
		ctx,
		azureProvider.Config.ResourceGroup,
		azureProvider.Config.MachineID+"-public-ip",
		nil,
	)
	if err != nil {
		return "", err
	}

	return *resource.Properties.IPAddress, nil
}
