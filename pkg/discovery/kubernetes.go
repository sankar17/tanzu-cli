// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"strings"

	"github.com/pkg/errors"

	configtypes "github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-cli/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-cli/pkg/cluster"
	"github.com/vmware-tanzu/tanzu-cli/pkg/common"
	"github.com/vmware-tanzu/tanzu-cli/pkg/distribution"
	"github.com/vmware-tanzu/tanzu-cli/pkg/utils"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/log"
)

// KubernetesDiscovery is an artifact discovery utilizing CLIPlugin API in kubernetes cluster
type KubernetesDiscovery struct {
	name           string
	kubeconfigPath string
	kubecontext    string
}

// NewKubernetesDiscovery returns a new kubernetes repository
func NewKubernetesDiscovery(name, kubeconfigPath, kubecontext string) Discovery {
	return &KubernetesDiscovery{
		name:           name,
		kubeconfigPath: kubeconfigPath,
		kubecontext:    kubecontext,
	}
}

// List available plugins.
func (k *KubernetesDiscovery) List() ([]Discovered, error) {
	return k.Manifest()
}

// Name of the repository.
func (k *KubernetesDiscovery) Name() string {
	return k.name
}

// Manifest returns the manifest for a kubernetes repository.
func (k *KubernetesDiscovery) Manifest() ([]Discovered, error) {
	// Create cluster client
	clusterClient, err := cluster.NewClient(k.kubeconfigPath, k.kubecontext, cluster.Options{})
	if err != nil {
		return nil, err
	}

	return k.GetDiscoveredPlugins(clusterClient)
}

// GetDiscoveredPlugins returns the list of discovered plugin from a kubernetes cluster
func (k *KubernetesDiscovery) GetDiscoveredPlugins(clusterClient cluster.Client) ([]Discovered, error) {
	plugins := make([]Discovered, 0)

	exists, err := clusterClient.VerifyCLIPluginCRD()
	if !exists || err != nil {
		logMsg := "Skipping context-aware plugin discovery because CLIPlugin CRD not present on the logged in cluster. "
		if err != nil {
			logMsg += err.Error()
		}
		log.V(4).Info(logMsg)
		return nil, nil
	}

	// get all cliplugins resources available on the cluster
	cliplugins, err := clusterClient.ListCLIPluginResources()
	if err != nil {
		return nil, err
	}

	imageRepositoryOverride, err := clusterClient.GetCLIPluginImageRepositoryOverride()
	if err != nil {
		log.Infof("unable to get image repository override information for some of the plugins. Error: %v", err)
	}

	// Convert all CLIPlugin resources to Discovered object
	for i := range cliplugins {
		dp, err := DiscoveredFromK8sV1alpha1WithImageRepositoryOverride(&cliplugins[i], imageRepositoryOverride)
		if err != nil {
			return nil, err
		}
		dp.Source = k.name
		dp.DiscoveryType = k.Type()
		plugins = append(plugins, dp)
	}

	return plugins, nil
}

// Type of the repository.
func (k *KubernetesDiscovery) Type() string {
	return common.DiscoveryTypeKubernetes
}

// DiscoveredFromK8sV1alpha1WithImageRepositoryOverride returns discovered plugin object from k8sV1alpha1
func DiscoveredFromK8sV1alpha1WithImageRepositoryOverride(p *cliv1alpha1.CLIPlugin, imageRepoOverride map[string]string) (Discovered, error) {
	// Update artifacts based on image repository override if applicable
	UpdateArtifactsBasedOnImageRepositoryOverride(p, imageRepoOverride)

	dp := Discovered{
		Name:               p.Name,
		Description:        p.Spec.Description,
		RecommendedVersion: p.Spec.RecommendedVersion,
		Optional:           p.Spec.Optional,
		Target:             configtypes.StringToTarget(string(p.Spec.Target)),
	}
	dp.SupportedVersions = make([]string, 0)
	for v := range p.Spec.Artifacts {
		dp.SupportedVersions = append(dp.SupportedVersions, v)
	}
	if err := utils.SortVersions(dp.SupportedVersions); err != nil {
		return dp, errors.Wrapf(err, "error parsing supported versions for plugin %s", p.Name)
	}

	dp.Distribution = distribution.ArtifactsFromK8sV1alpha1(p.Spec.Artifacts)
	return dp, nil
}

// UpdateArtifactsBasedOnImageRepositoryOverride updates artifacts based on image repository override
func UpdateArtifactsBasedOnImageRepositoryOverride(p *cliv1alpha1.CLIPlugin, imageRepoOverride map[string]string) {
	replaceImageRepository := func(a *cliv1alpha1.Artifact) {
		if a.Image != "" {
			for originalRepo, overrideRepo := range imageRepoOverride {
				if strings.HasPrefix(a.Image, originalRepo) {
					a.Image = strings.Replace(a.Image, originalRepo, overrideRepo, 1)
				}
			}
		}
	}
	for i := range p.Spec.Artifacts {
		for j := range p.Spec.Artifacts[i] {
			replaceImageRepository(&p.Spec.Artifacts[i][j])
		}
	}
}
