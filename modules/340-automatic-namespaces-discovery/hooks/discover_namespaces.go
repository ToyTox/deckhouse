/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hooks

import (
	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	discoverPatch = map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"extended-monitoring.flant.com/enabled": "true",
			},
		},
	}
	undiscoverPatch = map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"extended-monitoring.flant.com/enabled": nil,
			},
		},
	}
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	Queue: "/modules/automatic-namespaces-discovery/namespaces_discovery",
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:       "namespaces",
			ApiVersion: "v1",
			Kind:       "Namespace",
			FilterFunc: applyNamespaceFilter,
		},
	},
}, handleNamespace)

func applyNamespaceFilter(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	return obj.GetName(), nil
}

func handleNamespace(input *go_hook.HookInput) error {
	includeNames := input.Values.Get("automaticNamespaceDiscovery.includeNames").Array()
	excludeNames := input.Values.Get("automaticNamespaceDiscovery.excludeNames").Array()

	input.LogEntry.Infoln("includeNames:", includeNames) // TODO remove
	input.LogEntry.Infoln("excludeNames:", excludeNames) // TODO remove

	input.LogEntry.Infoln("Starting namespaces handler") // TODO remove
	snap := input.Snapshots["namespaces"]
	if len(snap) == 0 {
		input.LogEntry.Warnln("Namespaces not found. Skip")
		return nil
	}

	input.LogEntry.Infoln("Processing snapshot", snap) // TODO remove
NSLOOP:
	for _, ns := range snap {
		name := ns.(string)
		input.LogEntry.Infoln("Processing namespace:", name)
		for _, excludeName := range excludeNames {
			if excludeName.String() == name { // TODO enable pattern matching
				input.LogEntry.Infoln("Excluding namespace:", name)
				input.PatchCollector.MergePatch(undiscoverPatch, "v1", "Namespace", "", name)
				continue NSLOOP
			}
		}
		for _, includeName := range includeNames {
			if includeName.String() == name { // TODO enable pattern matching
				input.LogEntry.Infoln("Including namespace:", name)
				input.PatchCollector.MergePatch(discoverPatch, "v1", "Namespace", "", name)
				continue NSLOOP
			}
		}
	}

	input.LogEntry.Infoln("Exiting") // TODO remove
	return nil
}
