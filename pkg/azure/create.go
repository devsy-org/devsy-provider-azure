package azure

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/devsy-org/devsy-provider-azure/pkg/options"
	"github.com/devsy-org/devsy/pkg/ssh"
)

const adminUsername = "devsy"

func createVirtualNetwork(
	ctx context.Context,
	azureProvider *AzureProvider,
) (*armnetwork.VirtualNetwork, error) {
	vnet, exists := checkVirtualNetWork(ctx, azureProvider)
	if exists {
		return vnet, nil
	}

	vnetClient, err := armnetwork.NewVirtualNetworksClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.VirtualNetwork{
		Location: to.Ptr(azureProvider.Config.Zone),
		Tags:     azureProvider.Config.Tags,
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{
					to.Ptr("10.1.0.0/16"), // example 10.1.0.0/16
				},
			},
		},
	}

	pollerResponse, err := vnetClient.BeginCreateOrUpdate(
		ctx,
		azureProvider.Config.ResourceGroup,
		azureProvider.Config.MachineID+"-vnet",
		parameters,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.VirtualNetwork, nil
}

func createSubnets(ctx context.Context, azureProvider *AzureProvider) (*armnetwork.Subnet, error) {
	subnetClient, err := armnetwork.NewSubnetsClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr("10.1.10.0/24"),
		},
	}

	pollerResponse, err := subnetClient.BeginCreateOrUpdate(
		ctx,
		azureProvider.Config.ResourceGroup,
		azureProvider.Config.MachineID+"-vnet",
		azureProvider.Config.MachineID+"-subnet",
		parameters,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Subnet, nil
}

func createNetworkSecurityGroup(
	ctx context.Context,
	azureProvider *AzureProvider,
) (*armnetwork.SecurityGroup, error) {
	nsgClient, err := armnetwork.NewSecurityGroupsClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.SecurityGroup{
		Location: to.Ptr(azureProvider.Config.Zone),
		Tags:     azureProvider.Config.Tags,
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules: []*armnetwork.SecurityRule{
				sshRule("devsy_inbound_22", armnetwork.SecurityRuleDirectionInbound, "inbound"),
				sshRule("devsy_outbound_22", armnetwork.SecurityRuleDirectionOutbound, "outbound"),
			},
		},
	}

	pollerResponse, err := nsgClient.BeginCreateOrUpdate(
		ctx,
		azureProvider.Config.ResourceGroup,
		azureProvider.Config.MachineID+"-nsg",
		parameters,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SecurityGroup, nil
}

// normalizeCustomData ensures CustomData is base64 (reads + encodes if it's a path).
func normalizeCustomData(p *AzureProvider) error {
	data := p.Config.CustomData
	if data == "" {
		return nil
	}
	if _, err := os.Stat(data); err == nil {
		raw, err := os.ReadFile(data) //nolint:gosec // user-provided cloud-init path
		if err != nil {
			return err
		}
		p.Config.CustomData = base64.StdEncoding.EncodeToString(raw)
		return nil
	}
	if _, err := base64.StdEncoding.DecodeString(data); err != nil {
		return fmt.Errorf("custom data is not base64 encoded string or file")
	}
	return nil
}

// diskSizeGB casts to int32. Azure caps disk size at 32767 GB so overflow is impossible.
func diskSizeGB(v int) int32 {
	return int32(v) //nolint:gosec // bounded by Azure max 32767
}

func sshRule(
	name string,
	dir armnetwork.SecurityRuleDirection,
	label string,
) *armnetwork.SecurityRule {
	return &armnetwork.SecurityRule{
		Name: to.Ptr(name),
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			SourceAddressPrefix:      to.Ptr("0.0.0.0/0"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("0.0.0.0/0"),
			DestinationPortRange:     to.Ptr("22"),
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Priority:                 to.Ptr[int32](100),
			Description:              to.Ptr("devsy network security group " + label + " port 22"),
			Direction:                to.Ptr(dir),
		},
	}
}

func createPublicIP(
	ctx context.Context,
	azureProvider *AzureProvider,
) (*armnetwork.PublicIPAddress, error) {
	publicIPAddressClient, err := armnetwork.NewPublicIPAddressesClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.PublicIPAddress{
		Location: to.Ptr(azureProvider.Config.Zone),
		Tags:     azureProvider.Config.Tags,
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: to.Ptr(
				armnetwork.IPAllocationMethodStatic,
			), // Static or Dynamic
		},
	}

	pollerResponse, err := publicIPAddressClient.BeginCreateOrUpdate(
		ctx,
		azureProvider.Config.ResourceGroup,
		azureProvider.Config.MachineID+"-public-ip",
		parameters,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.PublicIPAddress, err
}

type networkInterfaceParams struct {
	subnetID               string
	publicIPID             string
	networkSecurityGroupID string
}

func createNetWorkInterface(
	ctx context.Context,
	azureProvider *AzureProvider,
	params networkInterfaceParams,
) (*armnetwork.Interface, error) {
	nicClient, err := armnetwork.NewInterfacesClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return nil, err
	}

	parameters := armnetwork.Interface{
		Location: to.Ptr(azureProvider.Config.Zone),
		Tags:     azureProvider.Config.Tags,
		Properties: &armnetwork.InterfacePropertiesFormat{
			// NetworkSecurityGroup:
			IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
				{
					Name: to.Ptr("ipConfig"),
					Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
						Subnet: &armnetwork.Subnet{
							ID: to.Ptr(params.subnetID),
						},
						PublicIPAddress: &armnetwork.PublicIPAddress{
							ID: to.Ptr(params.publicIPID),
						},
					},
				},
			},
			NetworkSecurityGroup: &armnetwork.SecurityGroup{
				ID: to.Ptr(params.networkSecurityGroupID),
			},
		},
	}

	pollerResponse, err := nicClient.BeginCreateOrUpdate(
		ctx,
		azureProvider.Config.ResourceGroup,
		azureProvider.Config.MachineID+"-nic",
		parameters,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Interface, err
}

func createVirtualMachine(
	ctx context.Context,
	azureProvider *AzureProvider,
	networkInterfaceID string,
) (*armcompute.VirtualMachine, error) {
	vmClient, err := armcompute.NewVirtualMachinesClient(
		azureProvider.Config.SubscriptionID,
		azureProvider.Cred,
		nil,
	)
	if err != nil {
		return nil, err
	}

	parameters, err := buildVirtualMachineParameters(azureProvider, networkInterfaceID)
	if err != nil {
		return nil, err
	}

	pollerResponse, err := vmClient.BeginCreateOrUpdate(
		ctx,
		azureProvider.Config.ResourceGroup,
		azureProvider.Config.MachineID,
		parameters,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.VirtualMachine, nil
}

func buildVirtualMachineParameters(
	azureProvider *AzureProvider,
	networkInterfaceID string,
) (armcompute.VirtualMachine, error) {
	publicKeyBase, err := ssh.GetPublicKeyBase(azureProvider.Config.MachineFolder)
	if err != nil {
		return armcompute.VirtualMachine{}, err
	}
	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase)
	if err != nil {
		return armcompute.VirtualMachine{}, err
	}
	if err := normalizeCustomData(azureProvider); err != nil {
		return armcompute.VirtualMachine{}, err
	}
	return armcompute.VirtualMachine{
		Location: to.Ptr(azureProvider.Config.Zone),
		Tags:     azureProvider.Config.Tags,
		Identity: &armcompute.VirtualMachineIdentity{
			Type: to.Ptr(armcompute.ResourceIdentityTypeSystemAssigned),
		},
		Properties: &armcompute.VirtualMachineProperties{
			StorageProfile: buildStorageProfile(azureProvider.Config),
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(
					armcompute.VirtualMachineSizeTypes(azureProvider.Config.MachineType),
				),
			},
			OSProfile:      buildOSProfile(azureProvider.Config, publicKey),
			NetworkProfile: buildNetworkProfile(networkInterfaceID),
		},
	}, nil
}

func buildStorageProfile(cfg *options.Options) *armcompute.StorageProfile {
	return &armcompute.StorageProfile{
		ImageReference: &armcompute.ImageReference{
			Offer:     to.Ptr(cfg.DiskImage.Offer),
			Publisher: to.Ptr(cfg.DiskImage.Publisher),
			SKU:       to.Ptr(cfg.DiskImage.SKU),
			Version:   to.Ptr(cfg.DiskImage.Version),
		},
		OSDisk: &armcompute.OSDisk{
			Name:         to.Ptr(cfg.MachineID + "-disk"),
			CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
			Caching:      to.Ptr(armcompute.CachingTypesReadWrite),
			ManagedDisk: &armcompute.ManagedDiskParameters{
				StorageAccountType: to.Ptr(armcompute.StorageAccountTypes(cfg.DiskType)),
			},
			DiskSizeGB: to.Ptr(diskSizeGB(cfg.DiskSizeGB)),
		},
	}
}

func buildOSProfile(cfg *options.Options, publicKey []byte) *armcompute.OSProfile {
	return &armcompute.OSProfile{
		ComputerName:  to.Ptr(cfg.MachineID),
		AdminUsername: to.Ptr(adminUsername),
		CustomData:    to.Ptr(cfg.CustomData),
		LinuxConfiguration: &armcompute.LinuxConfiguration{
			DisablePasswordAuthentication: to.Ptr(true),
			SSH: &armcompute.SSHConfiguration{
				PublicKeys: []*armcompute.SSHPublicKey{
					{
						Path:    to.Ptr("/home/" + adminUsername + "/.ssh/authorized_keys"),
						KeyData: to.Ptr(string(publicKey)),
					},
				},
			},
		},
	}
}

func buildNetworkProfile(networkInterfaceID string) *armcompute.NetworkProfile {
	return &armcompute.NetworkProfile{
		NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
			{ID: to.Ptr(networkInterfaceID)},
		},
	}
}
