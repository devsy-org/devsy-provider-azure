package options

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

var (
	AZURE_REGION          = "AZURE_REGION"
	AZURE_INSTANCE_SIZE   = "AZURE_INSTANCE_SIZE"
	AZURE_IMAGE           = "AZURE_IMAGE"
	AZURE_RESOURCE_GROUP  = "AZURE_RESOURCE_GROUP"
	AZURE_DISK_TYPE       = "AZURE_DISK_TYPE"
	AZURE_DISK_SIZE       = "AZURE_DISK_SIZE"
	AZURE_CUSTOM_DATA     = "AZURE_CUSTOM_DATA"
	AZURE_SUBSCRIPTION_ID = "AZURE_SUBSCRIPTION_ID"
	AZURE_TAGS            = "AZURE_TAGS"
)

type Options struct {
	DiskImage      AzureImage
	DiskSizeGB     int
	DiskType       string
	CustomData     string
	MachineFolder  string
	MachineID      string
	MachineType    string
	ResourceGroup  string
	SubscriptionID string
	Zone           string
	Tags           map[string]*string
}

type AzureImage struct {
	Offer     string
	Publisher string
	SKU       string
	Version   string
}

func FromEnv(init bool) (*Options, error) {
	retOptions := &Options{}
	if err := loadRequired(retOptions); err != nil {
		return nil, err
	}
	if err := loadDisk(retOptions); err != nil {
		return nil, err
	}
	if err := loadOptional(retOptions); err != nil {
		return nil, err
	}
	if init {
		return retOptions, nil
	}
	return loadMachine(retOptions)
}

func loadRequired(o *Options) error {
	for _, f := range []struct {
		key string
		dst *string
	}{
		{AZURE_RESOURCE_GROUP, &o.ResourceGroup},
		{AZURE_INSTANCE_SIZE, &o.MachineType},
		{AZURE_REGION, &o.Zone},
		{AZURE_SUBSCRIPTION_ID, &o.SubscriptionID},
	} {
		v, err := FromEnvOrError(f.key)
		if err != nil {
			return err
		}
		*f.dst = v
	}
	return nil
}

func loadDisk(o *Options) error {
	image, err := FromEnvOrError(AZURE_IMAGE)
	if err != nil {
		return err
	}
	parts := strings.Split(image, ":")
	if len(parts) < 4 {
		return fmt.Errorf("malformed image name")
	}
	o.DiskImage = AzureImage{
		Publisher: parts[0],
		Offer:     parts[1],
		SKU:       parts[2],
		Version:   parts[3],
	}

	diskSize, err := FromEnvOrError(AZURE_DISK_SIZE)
	if err != nil {
		return err
	}
	o.DiskSizeGB, err = strconv.Atoi(diskSize)
	if err != nil {
		return err
	}

	o.DiskType, err = FromEnvOrError(AZURE_DISK_TYPE)
	return err
}

func loadOptional(o *Options) error {
	o.CustomData = os.Getenv(AZURE_CUSTOM_DATA)
	tags, err := parseTags(os.Getenv(AZURE_TAGS))
	if err != nil {
		return err
	}
	o.Tags = tags
	return nil
}

func loadMachine(o *Options) (*Options, error) {
	id, err := FromEnvOrError("MACHINE_ID")
	if err != nil {
		return nil, err
	}
	o.MachineID = "devsy-" + id

	o.MachineFolder, err = FromEnvOrError("MACHINE_FOLDER")
	if err != nil {
		return nil, err
	}
	return o, nil
}

func FromEnvOrError(name string) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", fmt.Errorf(
			"couldn't find option %s in environment, please make sure %s is defined",
			name,
			name,
		)
	}

	return val, nil
}

func parseTags(tagsEnv string) (map[string]*string, error) {
	tags := map[string]*string{}
	if tagsEnv == "" {
		return tags, nil
	}

	for tag := range strings.SplitSeq(tagsEnv, ",") {
		key, value, ok := strings.Cut(tag, "=")
		if !ok {
			return tags, fmt.Errorf("malformed tag, expected format tagName=tagValue: %s", tag)
		}
		tags[key] = to.Ptr(value)
	}

	return tags, nil
}
