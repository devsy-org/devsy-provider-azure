package main

import (
	"bufio"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	providerName = "azure"
	githubOwner  = "devsy-org"
	githubRepo   = "devsy-provider-azure"
)

type Provider struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description"`
	Icon         string            `yaml:"icon"`
	IconDark     string            `yaml:"iconDark"`
	OptionGroups []OptionGroup     `yaml:"optionGroups"`
	Options      Options           `yaml:"options"`
	Agent        Agent             `yaml:"agent"`
	Binaries     Binaries          `yaml:"binaries"`
	Exec         map[string]string `yaml:"exec"`
}

type OptionGroup struct {
	Name           string   `yaml:"name"`
	DefaultVisible bool     `yaml:"defaultVisible"`
	Options        []string `yaml:"options"`
}

type Options map[string]Option

type Option struct {
	Description string   `yaml:"description,omitempty"`
	Required    bool     `yaml:"required,omitempty"`
	Default     string   `yaml:"default,omitempty"`
	Type        string   `yaml:"type,omitempty"`
	Suggestions []string `yaml:"suggestions,omitempty"`
	Command     string   `yaml:"command,omitempty"`
	Password    bool     `yaml:"password,omitempty"`
}

type Agent struct {
	Path                    string         `yaml:"path"`
	InactivityTimeout       string         `yaml:"inactivityTimeout"`
	InjectGitCredentials    string         `yaml:"injectGitCredentials"`
	InjectDockerCredentials string         `yaml:"injectDockerCredentials"`
	Binaries                map[string]any `yaml:"binaries"`
	Exec                    map[string]any `yaml:"exec"`
}

type Binaries struct {
	AzureProvider []Binary `yaml:"AZURE_PROVIDER"`
}

type Binary struct {
	OS       string `yaml:"os"`
	Arch     string `yaml:"arch"`
	Path     string `yaml:"path"`
	Checksum string `yaml:"checksum"`
}

type buildConfig struct {
	version     string
	projectRoot string
	isRelease   bool
	checksums   map[string]string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("expected version as argument")
	}

	cfg, err := newBuildConfig(os.Args[1])
	if err != nil {
		return err
	}

	provider := buildProvider(cfg)

	output, err := yaml.Marshal(provider)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}

	if _, err := os.Stdout.Write(output); err != nil {
		return fmt.Errorf("write yaml: %w", err)
	}
	return nil
}

func newBuildConfig(version string) (*buildConfig, error) {
	checksums, err := parseChecksums("./dist/checksums.txt")
	if err != nil {
		return nil, fmt.Errorf("parse checksums: %w", err)
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		owner := getEnvOrDefault("GITHUB_OWNER", githubOwner)
		projectRoot = fmt.Sprintf(
			"https://github.com/%s/%s/releases/download/%s",
			owner,
			githubRepo,
			version,
		)
	}

	isRelease := strings.Contains(projectRoot, "github.com") &&
		strings.Contains(projectRoot, "/releases/")

	return &buildConfig{
		version:     version,
		projectRoot: projectRoot,
		isRelease:   isRelease,
		checksums:   checksums,
	}, nil
}

func buildProvider(cfg *buildConfig) Provider {
	return Provider{
		Name:         providerName,
		Version:      cfg.version,
		Description:  "Devsy on Azure Cloud",
		Icon:         "https://raw.githubusercontent.com/devsy-org/devsy/main/desktop/src/images/azure.svg",
		IconDark:     "https://raw.githubusercontent.com/devsy-org/devsy/main/desktop/src/images/azure_white.svg",
		OptionGroups: buildOptionGroups(),
		Options:      buildOptions(),
		Agent:        buildAgent(cfg),
		Binaries:     buildBinaries(cfg, allPlatforms()),
		Exec: map[string]string{
			"init":    "${AZURE_PROVIDER} init",
			"command": "${AZURE_PROVIDER} command",
			"create":  "${AZURE_PROVIDER} create",
			"delete":  "${AZURE_PROVIDER} delete",
			"start":   "${AZURE_PROVIDER} start",
			"stop":    "${AZURE_PROVIDER} stop",
			"status":  "${AZURE_PROVIDER} status",
		},
	}
}

func buildOptionGroups() []OptionGroup {
	return []OptionGroup{
		{
			Name:           "Azure options",
			DefaultVisible: true,
			Options: []string{
				"AZURE_SUBSCRIPTION_ID",
				"AZURE_RESOURCE_GROUP",
				"AZURE_REGION",
				"AZURE_DISK_SIZE",
				"AZURE_DISK_TYPE",
				"AZURE_IMAGE",
				"AZURE_INSTANCE_SIZE",
				"AZURE_TAGS",
			},
		},
		{
			Name:           "Agent options",
			DefaultVisible: false,
			Options: []string{
				"AGENT_PATH",
				"INACTIVITY_TIMEOUT",
				"INJECT_DOCKER_CREDENTIALS",
				"INJECT_GIT_CREDENTIALS",
			},
		},
		{
			Name:           "Advanced options",
			DefaultVisible: false,
			Options:        []string{"AZURE_CUSTOM_DATA"},
		},
	}
}

func buildOptions() Options {
	opts := Options{}
	maps.Copy(opts, buildAzureOptions())
	maps.Copy(opts, buildAgentOptions())
	return opts
}

func buildAzureOptions() Options {
	return Options{
		"AZURE_SUBSCRIPTION_ID": {
			Description: "The azure subscription id",
			Required:    true,
			Command:     `az account show --query id --output tsv || true`,
		},
		"AZURE_RESOURCE_GROUP": {
			Description: "The azure resource group name",
			Required:    true,
			Command:     `printf "%s" "${AZURE_RESOURCE_GROUP:-$(az group list | jq '.[0].name' | tr -d '\"')}" || true`,
		},
		"AZURE_REGION": {
			Description: "The azure region to use",
			Required:    true,
			Command:     `printf "%s" "${AZURE_REGION:-}" || true`,
			Suggestions: azureRegions(),
		},
		"AZURE_DISK_SIZE": {
			Description: "The disk size to use.",
			Default:     "40",
		},
		"AZURE_IMAGE": {
			Description: "The disk image to use.",
			Default:     "Canonical:0001-com-ubuntu-server-jammy:22_04-lts-gen2:latest",
		},
		"AZURE_DISK_TYPE": {
			Description: "The disk type to use.",
			Default:     "StandardSSD_LRS",
			Suggestions: []string{
				"Standard_LRS", "StandardSSD_LRS", "StandardSSD_ZRS",
				"Premium_LRS", "PremiumV2_LRS", "Premium_ZRS",
			},
		},
		"AZURE_INSTANCE_SIZE": {
			Description: "The machine type to use.",
			Default:     "Standard_D4s_v3",
			Suggestions: azureInstanceSizes(),
		},
		"AZURE_CUSTOM_DATA": {
			Description: "The custom data to inject into the VM. E.g. cloud-init.txt or base64 string",
		},
		"AZURE_TAGS": {
			Description: "Extra tags to apply to all created resources. " +
				"Comma separated list, e.g. myTag=myvalue,myTag2=myValue2",
		},
	}
}

func buildAgentOptions() Options {
	return Options{
		"INACTIVITY_TIMEOUT": {
			Description: "If defined, will automatically stop the VM after the inactivity period.",
			Default:     "10m",
		},
		"INJECT_GIT_CREDENTIALS": {
			Description: "If Devsy should inject git credentials into the remote host.",
			Default:     "true",
		},
		"INJECT_DOCKER_CREDENTIALS": {
			Description: "If Devsy should inject docker credentials into the remote host.",
			Default:     "true",
		},
		"AGENT_PATH": {
			Description: "The path where to inject the Devsy agent to.",
			Default:     "/var/lib/toolbox/devsy",
		},
	}
}

func azureRegions() []string {
	return []string{
		"australiacentral", "australiaeast", "australiasoutheast", "brazilsouth",
		"canadacentral", "canadaeast", "centralindia", "centralus", "eastasia",
		"eastus", "eastus2", "francecentral", "germanywestcentral", "israelcentral",
		"italynorth", "japaneast", "japanwest", "jioindiawest", "koreacentral",
		"koreasouth", "northcentralus", "northeurope", "norwayeast", "polandcentral",
		"qatarcentral", "southafricanorth", "southcentralus", "southeastasia",
		"southindia", "swedencentral", "switzerlandnorth", "uaenorth", "uksouth",
		"ukwest", "westcentralus", "westeurope", "westindia", "westus", "westus2",
		"westus3",
	}
}

func azureInstanceSizes() []string {
	return []string{
		"Standard_B12ms", "Standard_B16ms", "Standard_B1ms", "Standard_B1s",
		"Standard_B20ms", "Standard_B2ms", "Standard_B2s", "Standard_B4ms",
		"Standard_B8ms", "Standard_D16s_v3", "Standard_D2s_v3", "Standard_D32s_v3",
		"Standard_D48s_v3", "Standard_D4s_v3", "Standard_D64s_v3", "Standard_D8s_v3",
		"Standard_DS1_v2", "Standard_DS2_v2", "Standard_DS3_v2", "Standard_DS4_v2",
		"Standard_DS5_v2", "Standard_E16s_v3", "Standard_E20s_v3", "Standard_E2s_v3",
		"Standard_E32s_v3", "Standard_E48s_v3", "Standard_E4s_v3", "Standard_E64s_v3",
		"Standard_E8s_v3", "Standard_F16s_v2", "Standard_F2s_v2", "Standard_F32s_v2",
		"Standard_F48s_v2", "Standard_F4s_v2", "Standard_F64s_v2", "Standard_F72s_v2",
		"Standard_F8s_v2",
	}
}

func buildAgent(cfg *buildConfig) Agent {
	return Agent{ //nolint:gosec // string values are shell var refs, not literal credentials
		Path:                    "${AGENT_PATH}",
		InactivityTimeout:       "${INACTIVITY_TIMEOUT}",
		InjectGitCredentials:    "${INJECT_GIT_CREDENTIALS}",
		InjectDockerCredentials: "${INJECT_DOCKER_CREDENTIALS}",
		Binaries: map[string]any{
			"AZURE_PROVIDER": buildBinaries(cfg, linuxPlatforms()).AzureProvider,
		},
		Exec: map[string]any{
			"shutdown": "${AZURE_PROVIDER} stop || shutdown",
		},
	}
}

func buildBinaries(cfg *buildConfig, platforms []string) Binaries {
	return Binaries{AzureProvider: buildBinaryList(cfg, platforms)}
}

func buildBinaryList(cfg *buildConfig, platforms []string) []Binary {
	result := make([]Binary, 0, len(platforms))
	for _, platform := range platforms {
		result = append(result, buildBinary(cfg, platform))
	}
	return result
}

func buildBinary(cfg *buildConfig, platform string) Binary {
	osName, arch, _ := strings.Cut(platform, "/")

	path := cfg.projectRoot
	if !cfg.isRelease {
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			base, _ := url.Parse(path)
			joined, _ := url.JoinPath(base.String(), buildDir(platform))
			path = joined
		} else {
			absPath, _ := filepath.Abs(path)
			path = filepath.Join(absPath, buildDir(platform))
		}
	}

	filename := fmt.Sprintf("devsy-provider-%s-%s-%s", providerName, osName, arch)
	if osName == "windows" {
		filename += ".exe"
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		path, _ = url.JoinPath(path, filename)
	} else {
		path = filepath.Join(path, filename)
	}

	return Binary{
		OS:       osName,
		Arch:     arch,
		Path:     path,
		Checksum: cfg.checksums[filename],
	}
}

const (
	platformLinuxAMD64   = "linux/amd64"
	platformLinuxARM64   = "linux/arm64"
	platformDarwinAMD64  = "darwin/amd64"
	platformDarwinARM64  = "darwin/arm64"
	platformWindowsAMD64 = "windows/amd64"
)

func buildDir(platform string) string {
	dirs := map[string]string{
		platformLinuxAMD64:   "build_linux_amd64_v1",
		platformLinuxARM64:   "build_linux_arm64_v8.0",
		platformDarwinAMD64:  "build_darwin_amd64_v1",
		platformDarwinARM64:  "build_darwin_arm64_v8.0",
		platformWindowsAMD64: "build_windows_amd64_v1",
	}
	return dirs[platform]
}

func allPlatforms() []string {
	return []string{
		platformLinuxAMD64, platformLinuxARM64,
		platformDarwinAMD64, platformDarwinARM64,
		platformWindowsAMD64,
	}
}

func linuxPlatforms() []string {
	return []string{platformLinuxAMD64, platformLinuxARM64}
}

func parseChecksums(path string) (map[string]string, error) {
	file, err := os.Open(path) //nolint:gosec // build-time tool reading goreleaser output
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	checksums := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if checksum, filename, ok := strings.Cut(scanner.Text(), " "); ok {
			checksums[strings.TrimSpace(filename)] = checksum
		}
	}

	return checksums, scanner.Err()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
