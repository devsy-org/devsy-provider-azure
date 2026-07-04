# Changelog

## [0.2.5](https://github.com/devsy-org/devsy-provider-azure/compare/v0.2.4...v0.2.5) (2026-07-04)


### Bug Fixes

* **deps:** update module github.com/devsy-org/devsy to v1.3.1 ([#35](https://github.com/devsy-org/devsy-provider-azure/issues/35)) ([b7df365](https://github.com/devsy-org/devsy-provider-azure/commit/b7df365f29456341293717b3bb04cbf56799537d))

## [0.2.4](https://github.com/devsy-org/devsy-provider-azure/compare/v0.2.3...v0.2.4) (2026-07-03)


### Bug Fixes

* **deps:** update azure-sdk-for-go monorepo ([#26](https://github.com/devsy-org/devsy-provider-azure/issues/26)) ([1225502](https://github.com/devsy-org/devsy-provider-azure/commit/12255023f9478ef0fc0e0fe3d16f530813773ac3))

## [0.2.3](https://github.com/devsy-org/devsy-provider-azure/compare/v0.2.2...v0.2.3) (2026-06-27)


### Bug Fixes

* **deps:** update azure-sdk-for-go monorepo ([#24](https://github.com/devsy-org/devsy-provider-azure/issues/24)) ([642e7dc](https://github.com/devsy-org/devsy-provider-azure/commit/642e7dc9bcfc95fd577377191738756aa6edd1ca))

## [0.2.2](https://github.com/devsy-org/devsy-provider-azure/compare/v0.2.1...v0.2.2) (2026-06-27)


### Bug Fixes

* **deps:** update module golang.org/x/crypto to v0.53.0 ([#22](https://github.com/devsy-org/devsy-provider-azure/issues/22)) ([0b8b5da](https://github.com/devsy-org/devsy-provider-azure/commit/0b8b5da7262b883492860e45811ee2b54520a8c1))

## [0.2.1](https://github.com/devsy-org/devsy-provider-azure/compare/v0.2.0...v0.2.1) (2026-06-27)


### Bug Fixes

* **deps:** update azure-sdk-for-go monorepo ([#20](https://github.com/devsy-org/devsy-provider-azure/issues/20)) ([d4b6f0f](https://github.com/devsy-org/devsy-provider-azure/commit/d4b6f0f6d3906a3824f5cfbec5f924640dbf26c3))

## [0.2.0](https://github.com/devsy-org/devsy-provider-azure/compare/v0.1.0...v0.2.0) (2026-06-26)


### Miscellaneous Chores

* release 0.2.0 ([#14](https://github.com/devsy-org/devsy-provider-azure/issues/14)) ([c1fe2c8](https://github.com/devsy-org/devsy-provider-azure/commit/c1fe2c89e20ea216895092fb0f82a27f2a64bebe))

## [0.1.0](https://github.com/devsy-org/devsy-provider-azure/compare/v0.0.1...v0.1.0) (2026-06-26)


### ⚠ BREAKING CHANGES

* AZURE_PROVIDER_TOKEN option, stop-remote and token subcommands have been removed.

### Features

* add AZURE_TAGS option to specify additional tags ([e33e1f7](https://github.com/devsy-org/devsy-provider-azure/commit/e33e1f7371de150dde30f9e1c65354c5f72c33ce))
* add AZURE_TAGS option to specify additional tags ([29306e5](https://github.com/devsy-org/devsy-provider-azure/commit/29306e51ea4f0ce34e1f4cae771bac9ababbedb5))
* add configurable disk type ([26f3c89](https://github.com/devsy-org/devsy-provider-azure/commit/26f3c89b8720b5814fbd1d42059e1b136dda337c))
* add configurable disk type ([edfda55](https://github.com/devsy-org/devsy-provider-azure/commit/edfda556a6655e4a0fcdf2e628e9ae8241efee08))
* migrate provider to devsy format ([c7013b4](https://github.com/devsy-org/devsy-provider-azure/commit/c7013b4013956552f2db5e121aa0ef5470490d64))
* migrate provider to devsy format ([7e64579](https://github.com/devsy-org/devsy-provider-azure/commit/7e64579c894103d5cedd93aa4270c0e5b6517587))
* support gen2 hypervisor, switch default image, improve region list, improve instance type list ([672110a](https://github.com/devsy-org/devsy-provider-azure/commit/672110aea59fe0c1d1c2226c40b4d756b7a1ee5e))
* support gen2 hypervisor, switch default image, improve region list, improve instance type list ([681a2b7](https://github.com/devsy-org/devsy-provider-azure/commit/681a2b7886e67629081e390f96ebcbbc3e09f8ad))


### Bug Fixes

* add missing brace to provider ([864d578](https://github.com/devsy-org/devsy-provider-azure/commit/864d578c89a76b861bdf4143b3c57683a40d4f6d))
* add missing region in manifest ([cc39e16](https://github.com/devsy-org/devsy-provider-azure/commit/cc39e16b18c30780961c349f2b82f1faa2dc06d9))
* **cd:** updat ego version in pipeline ([500656c](https://github.com/devsy-org/devsy-provider-azure/commit/500656c7ed9b6d5bf04a1a5a63dfccd675fdec6d))
* check vm response length before trying to find out current status ([757dfd4](https://github.com/devsy-org/devsy-provider-azure/commit/757dfd4cd216dbf21d32e9d7c7e1f7d0fc8cd882))
* **cmd:** add missing commands ([e4620ff](https://github.com/devsy-org/devsy-provider-azure/commit/e4620ff23c60a1b9f47d862b16d67398d53db0bf))
* escape provider commands ([3d4c27a](https://github.com/devsy-org/devsy-provider-azure/commit/3d4c27ad5215957f554939e29f480c4d673e6580))
* put corrent Azure icon in the provider ([656a8ba](https://github.com/devsy-org/devsy-provider-azure/commit/656a8ba5c5820cc5d508455935adfe8bdebadb74))
* put corrent Azure icon in the provider ([bac0755](https://github.com/devsy-org/devsy-provider-azure/commit/bac0755578ef337e95cc12e485e337a86d0f38e7))
* resolve golangci-lint findings ([1a80c38](https://github.com/devsy-org/devsy-provider-azure/commit/1a80c380d5f95304e1238f9dd5c2439c7287142f))
* retrieve credentials from SDK ([6233a5d](https://github.com/devsy-org/devsy-provider-azure/commit/6233a5d7ecd108c76175f797aa5cdeba43da3f72))
* retrieve credentials from SDK ([42e13bd](https://github.com/devsy-org/devsy-provider-azure/commit/42e13bdfe38e265e0f42432a97517b5e696bef7b))
* **stop:** use deallocation when stopping a machine ([f4883cb](https://github.com/devsy-org/devsy-provider-azure/commit/f4883cb4479a866d313db4b15ef65894dca41efd))
* **stop:** use deallocation when stopping a machine ([446f67b](https://github.com/devsy-org/devsy-provider-azure/commit/446f67b57317d8269cdfcf36c318ecabf805b955))
