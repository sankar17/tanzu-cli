// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	configtypes "github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-cli/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-cli/pkg/common"
	"github.com/vmware-tanzu/tanzu-cli/pkg/distribution"
	"github.com/vmware-tanzu/tanzu-cli/pkg/utils"
)

var defaultTimeout = 5 * time.Second

// Plugin contains information about a Tanzu CLI plugin discovered via a REST API.
type Plugin struct {
	// Name of the plugin.
	Name string `json:"name"`

	// Description is the plugin's description.
	Description string `json:"description"`

	// Recommended version that Tanzu CLI should use if available.
	// The value should be a valid semantic version as defined in
	// https://semver.org/. E.g., 2.0.1
	RecommendedVersion string `json:"recommendedVersion"`

	// Artifacts contains an artifact list for every supported version.
	Artifacts map[string]cliv1alpha1.ArtifactList `json:"artifacts"`

	// Optional specifies whether the plugin is mandatory or optional
	// If optional, the plugin will not get auto-downloaded as part of
	// `tanzu login` or `tanzu plugin sync` command
	// To view the list of plugin, user can use `tanzu plugin list` and
	// to download a specific plugin run, `tanzu plugin install <plugin-name>`
	Optional bool `json:"optional"`

	// Target the target of the plugin
	Target configtypes.Target `json:"target"`
}

// ListPluginsResponse defines the response from List Plugins API.
type ListPluginsResponse struct {
	Plugins []Plugin `json:"plugins"`
}

// RESTDiscovery is an artifact discovery utilizing CLIPlugin API in kubernetes cluster
type RESTDiscovery struct {
	// name of the discovery.
	name string
	// endpoint is the REST API server endpoint.
	// E.g., api.my-domain.local
	endpoint string
	// basePath is the base URL path of the plugin discovery API.
	// E.g., /v1alpha1/cli/plugins
	basePath string
	// client is the HTTP client used to make the REST API call.
	client *http.Client
}

// NewRESTDiscovery returns a new kubernetes repository
func NewRESTDiscovery(name, endpoint, basePath string) Discovery {
	return &RESTDiscovery{
		name:     name,
		endpoint: endpoint,
		basePath: basePath,
		client:   http.DefaultClient,
	}
}
func (d *RESTDiscovery) doRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	res, err := d.client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("API error, status code: %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

// List available plugins.
func (d *RESTDiscovery) List() ([]Discovered, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s", d.endpoint, d.basePath), http.NoBody)
	if err != nil {
		return nil, err
	}

	var res ListPluginsResponse
	if err := d.doRequest(req, &res); err != nil {
		return nil, err
	}

	// Convert all CLIPlugin resources to Discovered object
	plugins := make([]Discovered, 0)
	for i := range res.Plugins {
		dp, err := DiscoveredFromREST(&res.Plugins[i])
		if err != nil {
			return nil, err
		}
		if dp.Name == "" {
			continue
		}
		dp.Source = d.name
		plugins = append(plugins, dp)
	}

	return plugins, nil
}

// Name of the repository.
func (d *RESTDiscovery) Name() string {
	return d.name
}

// Type of the repository.
func (d *RESTDiscovery) Type() string {
	return common.DiscoveryTypeREST
}

// DiscoveredFromREST returns discovered plugin object from a REST API.
func DiscoveredFromREST(p *Plugin) (Discovered, error) {
	dp := Discovered{
		Name:               p.Name,
		Description:        p.Description,
		RecommendedVersion: p.RecommendedVersion,
		Optional:           p.Optional,
		Target:             configtypes.StringToTarget(string(p.Target)),
	}
	dp.SupportedVersions = make([]string, 0)
	for v := range p.Artifacts {
		dp.SupportedVersions = append(dp.SupportedVersions, v)
	}
	if err := utils.SortVersions(dp.SupportedVersions); err != nil {
		return dp, errors.Wrapf(err, "error parsing supported versions for plugin %s", p.Name)
	}
	dp.Distribution = distribution.ArtifactsFromK8sV1alpha1(p.Artifacts)
	return dp, nil
}
