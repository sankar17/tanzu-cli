// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	configtypes "github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"

	"github.com/vmware-tanzu/tanzu-cli/pkg/common"
	"github.com/vmware-tanzu/tanzu-cli/pkg/constants"
)

// Default Standalone Discovery configuration
// Value of this variables gets assigned during build time
var (
	// DefaultAllowedPluginRepositories this can be comma separated list of allowed registries
	DefaultAllowedPluginRepositories     = ""
	DefaultStandaloneDiscoveryRepository = ""
	DefaultStandaloneDiscoveryImagePath  = ""
	DefaultStandaloneDiscoveryImageTag   = ""
	DefaultStandaloneDiscoveryName       = "default"
	// DefaultStandaloneDiscoveryNameLocal Used for local discovery of sources.
	// Changing the default-local discovery source label to default and default will be used as a local discovery source
	// default and default-local will co-exist in the config.yaml i.e. If local discovery source is used and is now assigned the default name, the discovery source named default-local will still exist.
	// And recommend that it be manually removed from the config file.
	DefaultStandaloneDiscoveryNameLocal = "default"
	DefaultStandaloneDiscoveryType      = common.DistributionTypeOCI
	DefaultStandaloneDiscoveryLocalPath = ""
)

// CoreRepositoryName is the core repository name.
const CoreRepositoryName = "core"

// CoreBucketName is the name of the core plugin repository bucket to use.
var CoreBucketName = "tanzu-cli-framework"

// DefaultVersionSelector is to only use stable versions of plugins
const DefaultVersionSelector = configtypes.NoUnstableVersions

// DefaultEdition is the edition assumed when there is no value in the local config file
const DefaultEdition = "tkg"

// CoreGCPBucketRepository is the default GCP bucket repository.
var CoreGCPBucketRepository = configtypes.GCPPluginRepository{ // nolint:staticcheck // Deprecated
	BucketName: CoreBucketName,
	Name:       CoreRepositoryName,
}

// AdvancedRepositoryName is the advanced repository name.
const AdvancedRepositoryName = "advanced"

// AdvancedGCPBucketRepository is the GCP bucket repository for advanced plugins.
var AdvancedGCPBucketRepository = configtypes.GCPPluginRepository{ // nolint:staticcheck // Deprecated
	BucketName: "tanzu-cli-advanced-plugins",
	Name:       AdvancedRepositoryName,
}

// DefaultTMCPluginsArtifactRepository is the S3 bucket repository for TMC plugins.
const DefaultTMCPluginsArtifactRepository = "https://tmc-cli.s3-us-west-2.amazonaws.com/plugins/artifacts"

// DefaultRepositories are the default repositories for the CLI.
var DefaultRepositories = []configtypes.PluginRepository{ // nolint:staticcheck // Deprecated
	{
		GCPPluginRepository: &CoreGCPBucketRepository,
	},
}

// GetDefaultStandaloneDiscoveryImage returns the default Standalone Discovery image
// from the configured build time variables
func GetDefaultStandaloneDiscoveryImage() string {
	defaultStandaloneDiscoveryRepository := DefaultStandaloneDiscoveryRepository
	defaultStandaloneDiscoveryImagePath := DefaultStandaloneDiscoveryImagePath
	defaultStandaloneDiscoveryImageTag := DefaultStandaloneDiscoveryImageTag

	// Run-time overrides of the configuration
	if customImageRepo := os.Getenv(constants.ConfigVariableCustomImageRepository); customImageRepo != "" {
		defaultStandaloneDiscoveryRepository = customImageRepo
	}
	if imagePath := os.Getenv(constants.ConfigVariableDefaultStandaloneDiscoveryImagePath); imagePath != "" {
		defaultStandaloneDiscoveryImagePath = imagePath
	}
	if imageTag := os.Getenv(constants.ConfigVariableDefaultStandaloneDiscoveryImageTag); imageTag != "" {
		defaultStandaloneDiscoveryImageTag = imageTag
	}

	return strings.Trim(defaultStandaloneDiscoveryRepository, "/") + "/" + strings.Trim(defaultStandaloneDiscoveryImagePath, "/") + ":" + defaultStandaloneDiscoveryImageTag
}

// GetDefaultStandaloneDiscoveryType returns the default standalone discovery type
func GetDefaultStandaloneDiscoveryType() string {
	// Run-time overrides of the configuration
	if dType := os.Getenv(constants.ConfigVariableDefaultStandaloneDiscoveryType); dType != "" {
		return dType
	}
	return DefaultStandaloneDiscoveryType
}

// GetDefaultStandaloneDiscoveryLocalPath returns default standalone discovery local path
func GetDefaultStandaloneDiscoveryLocalPath() string {
	// Run-time overrides of the configuration
	if localPath := os.Getenv(constants.ConfigVariableDefaultStandaloneDiscoveryLocalPath); localPath != "" {
		return localPath
	}
	return DefaultStandaloneDiscoveryLocalPath
}

// GetTrustedRegistries returns the list of trusted registries that can be used for
// downloading the CLIPlugins
func GetTrustedRegistries() []string {
	var trustedRegistries []string

	// Add default allowed registries to trusted registries
	if DefaultAllowedPluginRepositories != "" {
		for _, r := range strings.Split(DefaultAllowedPluginRepositories, ",") {
			trustedRegistries = append(trustedRegistries, strings.TrimSpace(r))
		}
	}

	// If custom image repository is defined add it to the list of trusted registries
	if customImageRepo := os.Getenv(constants.ConfigVariableCustomImageRepository); customImageRepo != "" {
		trustedRegistries = append(trustedRegistries, customImageRepo)
	}

	// If the pre-release plugin repo variable is set, add its host to the list of trusted registries
	if preReleaseRepoImage := os.Getenv(constants.ConfigVariablePreReleasePluginRepoImage); preReleaseRepoImage != "" {
		if u, err := url.ParseRequestURI("https://" + preReleaseRepoImage); err == nil {
			trustedRegistries = append(trustedRegistries, u.Hostname())
		}
	}

	// Add default central plugin discovery image to the trusted registries
	if u, err := url.ParseRequestURI("https://" + constants.TanzuCLIDefaultCentralPluginDiscoveryImage); err == nil {
		trustedRegistries = append(trustedRegistries, u.Hostname())
	}

	// Add any additional test central plugin discovery images to the trusted registries
	testDiscoveryImages := GetAdditionalTestDiscoveryImages()
	for _, image := range testDiscoveryImages {
		if u, err := url.ParseRequestURI("https://" + image); err == nil {
			trustedRegistries = append(trustedRegistries, u.Hostname())
		}
	}

	// If ALLOWED_REGISTRY environment variable is specified, allow those registries as well
	if allowedRegistry := os.Getenv(constants.AllowedRegistries); allowedRegistry != "" {
		for _, r := range strings.Split(allowedRegistry, ",") {
			trustedRegistries = append(trustedRegistries, strings.TrimSpace(r))
		}
	}

	return trustedRegistries
}

func GetAdditionalTestDiscoveryImages() []string {
	var additionalImages []string
	testDiscoveryImages := strings.Split(os.Getenv(constants.ConfigVariableAdditionalDiscoveryForTesting), ",")
	for _, image := range testDiscoveryImages {
		image = strings.TrimSpace(image)
		if image != "" {
			additionalImages = append(additionalImages, image)
		}
	}
	return additionalImages
}

func getHTTPURIForGCPPluginRepository(repo configtypes.GCPPluginRepository) string { // nolint:staticcheck // Deprecated
	return fmt.Sprintf("https://storage.googleapis.com/%s/", repo.BucketName)
}

// GetTrustedArtifactLocations returns the list of trusted URI prefixes that can
// be trusted for downloading the CLIPlugins. Currently, this includes only the
// "tanzu-cli-advanced-plugins" GCP bucket where TMC plugins are stored. Other
// exceptions can be added as and when necessary.
func GetTrustedArtifactLocations() []string {
	trustedLocations := []string{
		getHTTPURIForGCPPluginRepository(AdvancedGCPBucketRepository),
		DefaultTMCPluginsArtifactRepository,
	}

	return trustedLocations
}
