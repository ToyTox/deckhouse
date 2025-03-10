// Copyright 2021 Flant JSC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/peterbourgon/mergemap"
	"sigs.k8s.io/yaml"

	"github.com/deckhouse/deckhouse/dhctl/pkg/app"
	"github.com/deckhouse/deckhouse/dhctl/pkg/log"
)

type MetaConfig struct {
	ClusterType          string                 `json:"-"`
	Layout               string                 `json:"-"`
	ProviderName         string                 `json:"-"`
	OriginalProviderName string                 `json:"-"`
	ClusterPrefix        string                 `json:"-"`
	ClusterDNSAddress    string                 `json:"-"`
	DeckhouseConfig      DeckhouseClusterConfig `json:"-"`
	MasterNodeGroupSpec  MasterNodeGroupSpec    `json:"-"`
	TerraNodeGroupSpecs  []TerraNodeGroupSpec   `json:"-"`

	ClusterConfig     map[string]json.RawMessage `json:"clusterConfiguration"`
	InitClusterConfig map[string]json.RawMessage `json:"-"`

	ProviderClusterConfig map[string]json.RawMessage `json:"providerClusterConfiguration,omitempty"`
	StaticClusterConfig   map[string]json.RawMessage `json:"staticClusterConfiguration,omitempty"`

	VersionMap map[string]interface{} `json:"-"`
	Images     ImagesTags             `json:"-"`
	Registry   RegistryData           `json:"-"`
	UUID       string                 `json:"clusterUUID,omitempty"`
}

type ImagesTags map[string]map[string]interface{}

type RegistryData struct {
	Address   string `json:"address"`
	Path      string `json:"path"`
	Scheme    string `json:"scheme"`
	CA        string `json:"ca"`
	DockerCfg string `json:"dockerCfg"`
}

// Prepare extracts all necessary information from raw json messages to the root structure
func (m *MetaConfig) Prepare() (*MetaConfig, error) {
	if len(m.ClusterConfig) > 0 {
		if err := json.Unmarshal(m.ClusterConfig["clusterType"], &m.ClusterType); err != nil {
			return nil, fmt.Errorf("unable to parse cluster type from cluster configuration: %v", err)
		}

		var serviceSubnet string
		if err := json.Unmarshal(m.ClusterConfig["serviceSubnetCIDR"], &serviceSubnet); err != nil {
			return nil, fmt.Errorf("unable to unmarshal service subnet CIDR from cluster configuration: %v", err)
		}
		m.ClusterDNSAddress = getDNSAddress(serviceSubnet)
	}

	if len(m.InitClusterConfig) > 0 {
		if err := json.Unmarshal(m.InitClusterConfig["deckhouse"], &m.DeckhouseConfig); err != nil {
			return nil, fmt.Errorf("unable to unmarshal deckhouse configuration: %v", err)
		}

		m.Registry.DockerCfg = m.DeckhouseConfig.RegistryDockerCfg
		m.Registry.Scheme = strings.ToLower(m.DeckhouseConfig.RegistryScheme)
		m.Registry.CA = m.DeckhouseConfig.RegistryCA

		parts := strings.SplitN(m.DeckhouseConfig.ImagesRepo, "/", 2)
		m.Registry.Address = parts[0]
		if len(parts) == 2 {
			m.Registry.Path = fmt.Sprintf("/%s", parts[1])
		}
	}

	if m.ClusterType != CloudClusterType || len(m.ProviderClusterConfig) == 0 {
		return m, nil
	}

	if err := json.Unmarshal(m.ProviderClusterConfig["layout"], &m.Layout); err != nil {
		return nil, fmt.Errorf("unable to unmarshal layout from cluster configuration: %v", err)
	}
	m.Layout = strcase.ToKebab(m.Layout)

	var cloud ClusterConfigCloudSpec
	if err := json.Unmarshal(m.ClusterConfig["cloud"], &cloud); err != nil {
		return nil, fmt.Errorf("unable to unmarshal cloud section from provider cluster configuration: %v", err)
	}

	m.ProviderName = strings.ToLower(cloud.Provider)
	m.OriginalProviderName = cloud.Provider
	m.ClusterPrefix = cloud.Prefix

	if err := json.Unmarshal(m.ProviderClusterConfig["masterNodeGroup"], &m.MasterNodeGroupSpec); err != nil {
		return nil, fmt.Errorf("unable to unmarshal master node group from provider cluster configuration: %v", err)
	}

	m.TerraNodeGroupSpecs = []TerraNodeGroupSpec{}
	nodeGroups, ok := m.ProviderClusterConfig["nodeGroups"]
	if ok {
		if err := json.Unmarshal(nodeGroups, &m.TerraNodeGroupSpecs); err != nil {
			return nil, fmt.Errorf("unable to unmarshal static nodes from provider cluster configuration: %v", err)
		}
	}

	m.Registry.DockerCfg = m.DeckhouseConfig.RegistryDockerCfg
	m.Registry.Scheme = strings.ToLower(m.DeckhouseConfig.RegistryScheme)
	m.Registry.CA = m.DeckhouseConfig.RegistryCA

	parts := strings.SplitN(m.DeckhouseConfig.ImagesRepo, "/", 2)
	m.Registry.Address = parts[0]
	if len(parts) == 2 {
		m.Registry.Path = fmt.Sprintf("/%s", parts[1])
	}

	return m, nil
}

// MergeDeckhouseConfig returns deckhouse config merged from different sources
func (m *MetaConfig) MergeDeckhouseConfig(configs ...[]byte) map[string]interface{} {
	deckhouseModuleConfig := map[string]interface{}{
		"logLevel": m.DeckhouseConfig.LogLevel,
		"bundle":   m.DeckhouseConfig.Bundle,
	}

	if m.DeckhouseConfig.ReleaseChannel != "" {
		deckhouseModuleConfig["releaseChannel"] = m.DeckhouseConfig.ReleaseChannel
	}

	baseDeckhouseConfig := map[string]interface{}{"deckhouse": deckhouseModuleConfig}
	if len(configs) == 0 {
		return mergemap.Merge(baseDeckhouseConfig, m.DeckhouseConfig.ConfigOverrides)
	}

	var firstConfig map[string]interface{}
	_ = json.Unmarshal(configs[0], &firstConfig)

	for _, configRaw := range configs[1:] {
		var config map[string]interface{}
		_ = json.Unmarshal(configRaw, &config)

		firstConfig = mergemap.Merge(firstConfig, config)
	}

	firstConfig = mergemap.Merge(firstConfig, m.DeckhouseConfig.ConfigOverrides)
	firstConfig = mergemap.Merge(firstConfig, baseDeckhouseConfig)

	return firstConfig
}

func (m *MetaConfig) GetTerraNodeGroups() []TerraNodeGroupSpec {
	return m.TerraNodeGroupSpecs
}

func (m *MetaConfig) FindTerraNodeGroup(nodeGroupName string) []byte {
	for index, ng := range m.TerraNodeGroupSpecs {
		if ng.Name == nodeGroupName {
			var terraNodeGroups []json.RawMessage
			err := json.Unmarshal(m.ProviderClusterConfig["nodeGroups"], &terraNodeGroups)
			if err != nil {
				log.ErrorLn(err)
				return nil
			}
			return terraNodeGroups[index]
		}
	}
	return nil
}

func (m *MetaConfig) ExtractMasterNodeGroupStaticSettings() map[string]interface{} {
	static := make(map[string]interface{})

	if len(m.StaticClusterConfig) == 0 {
		return static
	}

	var internalNetworkCIDRs []string
	if data, ok := m.StaticClusterConfig["internalNetworkCIDRs"]; ok {
		err := json.Unmarshal(data, &internalNetworkCIDRs)
		if err != nil {
			log.DebugF("unmarshalling internalNetworkCIDRs: %v\n", err)
			return static
		}
	}

	static["internalNetworkCIDRs"] = internalNetworkCIDRs
	return static
}

// MasterNodeGroupManifest prepares NodeGroup custom resource for master nodes
func (m *MetaConfig) MasterNodeGroupManifest() map[string]interface{} {
	spec := map[string]interface{}{
		"nodeType": "CloudPermanent",
		"disruptions": map[string]interface{}{
			"approvalMode": "Manual",
		},
		"nodeTemplate": map[string]interface{}{
			"labels": map[string]interface{}{
				"node-role.kubernetes.io/master":        "",
				"node-role.kubernetes.io/control-plane": "",
			},
			"taints": []map[string]interface{}{
				{
					"key":    "node-role.kubernetes.io/master",
					"effect": "NoSchedule",
				},
			},
		},
	}
	if m.ClusterType == StaticClusterType {
		spec["nodeType"] = "Static"
	}

	return map[string]interface{}{
		"apiVersion": "deckhouse.io/v1",
		"kind":       "NodeGroup",
		"metadata": map[string]interface{}{
			"name": "master",
		},
		"spec": spec,
	}
}

// NodeGroupManifest prepares NodeGroup custom resource for static nodes, which were ordered by Terraform
func (m *MetaConfig) NodeGroupManifest(terraNodeGroup TerraNodeGroupSpec) map[string]interface{} {
	if terraNodeGroup.NodeTemplate == nil {
		terraNodeGroup.NodeTemplate = make(map[string]interface{})
	}
	return map[string]interface{}{
		"apiVersion": "deckhouse.io/v1",
		"kind":       "NodeGroup",
		"metadata": map[string]interface{}{
			"name": terraNodeGroup.Name,
		},
		"spec": map[string]interface{}{
			"nodeType": "CloudPermanent",
			"disruptions": map[string]interface{}{
				"approvalMode": "Manual",
			},
			"nodeTemplate": terraNodeGroup.NodeTemplate,
		},
	}
}

func (m *MetaConfig) MarshalFullConfig() []byte {
	data, _ := json.Marshal(m)
	return data
}

func (m *MetaConfig) MarshalConfig() []byte {
	newM := m.DeepCopy()
	newM.StaticClusterConfig = nil
	data, _ := json.Marshal(newM)
	return data
}

func (m *MetaConfig) ClusterConfigYAML() ([]byte, error) {
	if m.ClusterConfig == nil {
		return []byte{}, nil
	}
	return yaml.Marshal(m.ClusterConfig)
}

func (m *MetaConfig) ProviderClusterConfigYAML() ([]byte, error) {
	if m.ProviderClusterConfig == nil {
		return []byte{}, nil
	}
	return yaml.Marshal(m.ProviderClusterConfig)
}

func (m *MetaConfig) StaticClusterConfigYAML() ([]byte, error) {
	if m.StaticClusterConfig == nil {
		return []byte{}, nil
	}
	return yaml.Marshal(m.StaticClusterConfig)
}

func (m *MetaConfig) ConfigForKubeadmTemplates(nodeIP string) (map[string]interface{}, error) {
	data := make(map[string]interface{}, len(m.ClusterConfig))

	for key, value := range m.ClusterConfig {
		var t interface{}
		err := json.Unmarshal(value, &t)
		if err != nil {
			return nil, fmt.Errorf("cluster config unmarshal: %v", err)
		}
		data[key] = t
	}

	result := make(map[string]interface{})
	for key, value := range m.VersionMap {
		result[key] = value
	}

	result["runType"] = "ClusterBootstrap"
	result["extraArgs"] = make(map[string]interface{})
	result["clusterConfiguration"] = data
	// bashible will use this as a placeholder on envsubst call, address will be discovered in one of bashible steps
	result["nodeIP"] = "$MY_IP"

	if nodeIP != "" {
		result["nodeIP"] = nodeIP
	}

	registryData, err := m.ParseRegistryData()
	if err != nil {
		return nil, err
	}

	result["registry"] = registryData

	images := m.Images

	if !app.DontUsePublicControlPlaneImages {
		k8s := strings.Replace(fmt.Sprintf("%s", data["kubernetesVersion"]), ".", "", 1)
		delete(images["controlPlaneManager"], "etcd")
		delete(images["controlPlaneManager"], "kubeApiserver"+k8s)
		delete(images["controlPlaneManager"], "kubeControllerManager"+k8s)
		delete(images["controlPlaneManager"], "kubeScheduler"+k8s)
	}

	result["images"] = images.ConvertToMap()
	return result, nil
}

func (m *MetaConfig) ConfigForBashibleBundleTemplate(bundle, nodeIP string) (map[string]interface{}, error) {
	data := make(map[string]interface{}, len(m.ClusterConfig))

	for key, value := range m.ClusterConfig {
		var t interface{}
		err := json.Unmarshal(value, &t)
		if err != nil {
			return nil, fmt.Errorf("cluster config unmarshal: %v", err)
		}
		data[key] = t
	}

	clusterBootstrap := map[string]interface{}{
		"clusterDomain":     data["clusterDomain"],
		"clusterDNSAddress": m.ClusterDNSAddress,
	}

	if nodeIP != "" {
		clusterBootstrap["cloud"] = map[string]interface{}{"nodeIP": nodeIP}
	}

	nodeGroup := map[string]interface{}{
		"name":     "master",
		"nodeType": "CloudPermanent",
		"cloudInstances": map[string]interface{}{
			"classReference": map[string]string{
				"name": "master",
			},
		},
	}

	if m.ClusterType == StaticClusterType {
		nodeGroup["nodeType"] = "Static"
		nodeGroup["static"] = m.ExtractMasterNodeGroupStaticSettings()
	}

	registryData, err := m.ParseRegistryData()
	if err != nil {
		return nil, err
	}

	configForBashibleBundleTemplate := make(map[string]interface{})
	for key, value := range m.VersionMap {
		configForBashibleBundleTemplate[key] = value
	}

	configForBashibleBundleTemplate["runType"] = "ClusterBootstrap"
	configForBashibleBundleTemplate["bundle"] = bundle
	configForBashibleBundleTemplate["cri"] = data["defaultCRI"]
	configForBashibleBundleTemplate["kubernetesVersion"] = data["kubernetesVersion"]
	configForBashibleBundleTemplate["nodeGroup"] = nodeGroup
	configForBashibleBundleTemplate["clusterBootstrap"] = clusterBootstrap
	configForBashibleBundleTemplate["packagesProxy"] = make(map[string]interface{})
	if data["packagesProxy"] != nil {
		configForBashibleBundleTemplate["packagesProxy"] = data["packagesProxy"]
	}
	configForBashibleBundleTemplate["registry"] = registryData

	images := m.Images
	if !app.DontUsePublicControlPlaneImages {
		delete(images["common"], "pause")
	}
	configForBashibleBundleTemplate["images"] = images.ConvertToMap()

	return configForBashibleBundleTemplate, nil
}

// NodeGroupConfig returns values for terraform to order master node or static node
func (m *MetaConfig) NodeGroupConfig(nodeGroupName string, nodeIndex int, cloudConfig string) []byte {
	result := map[string]interface{}{
		"clusterConfiguration":         m.ClusterConfig,
		"providerClusterConfiguration": m.ProviderClusterConfig,
		"nodeIndex":                    nodeIndex,
		"cloudConfig":                  cloudConfig,
	}

	if nodeGroupName != "master" {
		result["nodeGroupName"] = nodeGroupName
	}

	if len(m.UUID) > 0 {
		result["clusterUUID"] = m.UUID
	}

	data, _ := json.Marshal(result)
	return data
}

func (m *MetaConfig) CachePath() string {
	return fmt.Sprintf("%s-%s-terraform-state-cache", m.ClusterPrefix, m.ProviderName)
}

func (m *MetaConfig) DeepCopy() *MetaConfig {
	out := MetaConfig{}

	if m.ClusterConfig != nil {
		config := make(map[string]json.RawMessage, len(m.ClusterConfig))
		for k, v := range m.ClusterConfig {
			config[k] = v
		}
		out.ClusterConfig = config
	}

	if m.InitClusterConfig != nil {
		config := make(map[string]json.RawMessage, len(m.InitClusterConfig))
		for k, v := range m.InitClusterConfig {
			config[k] = v
		}
		out.InitClusterConfig = config
	}

	if m.ProviderClusterConfig != nil {
		config := make(map[string]json.RawMessage, len(m.ProviderClusterConfig))
		for k, v := range m.ProviderClusterConfig {
			config[k] = v
		}
		out.ProviderClusterConfig = config
	}

	if m.StaticClusterConfig != nil {
		config := make(map[string]json.RawMessage, len(m.StaticClusterConfig))
		for k, v := range m.StaticClusterConfig {
			config[k] = v
		}
		out.StaticClusterConfig = config
	}

	out.Registry = m.Registry

	if m.ClusterType != "" {
		out.ClusterType = m.ClusterType
	}

	if m.ClusterPrefix != "" {
		out.ClusterPrefix = m.ClusterPrefix
	}

	if m.Layout != "" {
		out.Layout = m.Layout
	}

	if m.ProviderName != "" {
		out.ProviderName = m.ProviderName
	}

	if m.OriginalProviderName != "" {
		out.OriginalProviderName = m.OriginalProviderName
	}

	if m.UUID != "" {
		out.UUID = m.UUID
	}

	return m
}

func (m *MetaConfig) LoadVersionMap(filename string) error {
	versionMap := make(map[string]interface{})

	versionMapFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%s file load: %v", filename, err)
	}

	err = yaml.Unmarshal(versionMapFile, &versionMap)
	if err != nil {
		return fmt.Errorf("%s file unmarshal: %v", filename, err)
	}

	m.VersionMap = versionMap

	return nil
}

func (m *MetaConfig) ParseRegistryData() (map[string]interface{}, error) {
	type dockerCfg struct {
		Auths map[string]struct {
			Auth string `json:"auth"`
		} `json:"auths"`
	}

	var (
		registryAuth string
		dc           dockerCfg
	)

	log.DebugF("registry data: %v\n", m.Registry)

	if m.Registry.DockerCfg != "" {
		bytes, err := base64.StdEncoding.DecodeString(m.Registry.DockerCfg)
		if err != nil {
			return nil, fmt.Errorf("cannot base64 decode docker cfg: %v", err)
		}

		log.DebugF("parse registry data: dockerCfg after base64 decode = %s\n", bytes)
		err = json.Unmarshal(bytes, &dc)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal docker cfg: %v", err)
		}

		if registry, ok := dc.Auths[m.Registry.Address]; ok {
			registryAuth = registry.Auth
		}
	}

	ret := m.Registry.ConvertToMap()
	ret["auth"] = registryAuth

	return ret, nil
}

func (m *MetaConfig) LoadImagesTags(filename string) error {
	var imagesTags ImagesTags

	imagesTagsJSONFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%s file load: %v", filename, err)
	}

	err = yaml.Unmarshal(imagesTagsJSONFile, &imagesTags)
	if err != nil {
		return fmt.Errorf("%s file unmarshal: %v", filename, err)
	}

	m.Images = imagesTags

	return nil
}

func (r *RegistryData) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"address":   r.Address,
		"path":      r.Path,
		"scheme":    r.Scheme,
		"ca":        r.CA,
		"dockerCfg": r.DockerCfg,
	}
}

func getDNSAddress(serviceCIDR string) string {
	ip, ipnet, err := net.ParseCIDR(serviceCIDR)
	if err != nil {
		log.DebugLn("serviceSubnetCIDR is not valid CIDR (should be validated with openapi scheme)")
		return ""
	}

	inc := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

	clusterDNS := ""
	counter := 0
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		// The .10 address for /24 network is dns address
		if counter == 10 {
			clusterDNS = ip.String()
			break
		}
		counter++
	}

	return clusterDNS
}

func (i *ImagesTags) ConvertToMap() map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range *i {
		res[k] = v
	}
	return res
}
