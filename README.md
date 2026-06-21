# Azure Provider for Devsy

[![Open in Devsy!](https://img.shields.io/badge/open_in_devsy-8A2BE2?style=for-the-badge)](https://devsy.sh/open#https://github.com/devsy-org/devsy-provider-azure)

## Getting started

The provider is available for auto-installation using

```sh
devsy provider add azure
devsy provider use azure
```

Follow the on-screen instructions to complete the setup.

Needed variables will be:

- AZURE_SUBSCRIPTION_ID
- AZURE_RESOURCE_GROUP
- AZURE_REGION

Authentication is obtained using Azure's Default Credential authenticator, which uses
the CLI tool, the environment, or certificates. See
[the Azure CLI auth docs](https://learn.microsoft.com/en-us/cli/azure/authenticate-azure-cli)
for setup options.

### Required permissions

The identity running `devsy up`/`create` needs to provision the VM, its
networking resources, and a role assignment that lets the VM stop itself
when the inactivity timer fires. The simplest setup is **Owner** (or
**Contributor** + **User Access Administrator**) on the target resource
group. Specifically the provider needs:

- `Microsoft.Compute/virtualMachines/*` (VM lifecycle)
- `Microsoft.Network/*` (vnet, subnet, NSG, public IP, NIC)
- `Microsoft.Authorization/roleAssignments/write` — used once per VM to
  grant the VM's system-assigned managed identity the **Virtual Machine
  Contributor** role scoped to itself, so `${AZURE_PROVIDER} stop`
  invoked from inside the VM can deallocate.

### Creating your first devsy env with azure

After the initial setup, just use:

```sh
devsy up .
```

You'll need to wait for the machine and environment setup.

### Customize the VM Instance

This provider has the following options:

| NAME                  | REQUIRED | DESCRIPTION                              | DEFAULT                                                      |
| --------------------- | -------- | ---------------------------------------- | ------------------------------------------------------------ |
| AZURE_SUBSCRIPTION_ID | true     | The azure subscription id                |                                                              |
| AZURE_RESOURCE_GROUP  | true     | The azure resource group name            |                                                              |
| AZURE_REGION          | true     | The azure region to use                  |                                                              |
| AZURE_INSTANCE_SIZE   | false    | The machine type to use.                 | Standard_D4s_v3                                              |
| AZURE_IMAGE           | false    | The disk image to use.                   | Canonical:0001-com-ubuntu-server-jammy:22_04-lts-gen2:latest |
| AZURE_DISK_SIZE       | false    | The disk size in GB.                     | 40                                                           |
| AZURE_DISK_TYPE       | false    | The disk type to use.                    | StandardSSD_LRS                                              |
| AZURE_CUSTOM_DATA     | false    | Cloud-init file or base64-encoded string |                                                              |
| AZURE_TAGS            | false    | Comma-separated `key=value` tags         |                                                              |

Options can be set in `env` or via the CLI:

```sh
devsy provider set-options -o AZURE_IMAGE=Vendor:Offer:SKU:Version
```

## Local Development

To build and test the provider locally, use [task](https://taskfile.dev/) `task build:provider:dev`. The provider file is created in `./dist/provider.yaml`.
