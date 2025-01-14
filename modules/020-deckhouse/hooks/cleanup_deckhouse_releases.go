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
	"sort"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"
	"github.com/flant/shell-operator/pkg/kube/object_patch"

	"github.com/deckhouse/deckhouse/modules/020-deckhouse/hooks/internal/v1alpha1"
)

/*
  This hook handle invalid situation when more then 1 Deployed release exists at the moment:
    Hook move all releases except the latest one to the Outdated state

  The hook will keep only 10 Outdated releases, removing others
*/

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	Queue: "/modules/deckhouse/cleanup_deckhouse_release",
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:       "releases",
			ApiVersion: "deckhouse.io/v1alpha1",
			Kind:       "DeckhouseRelease",
			FilterFunc: filterDeckhouseRelease,
		},
	},
}, cleanupReleases)

func cleanupReleases(input *go_hook.HookInput) error {
	snap := input.Snapshots["releases"]
	if len(snap) == 0 {
		return nil
	}

	releases := make([]deckhouseRelease, 0, len(snap))
	for _, sn := range snap {
		releases = append(releases, sn.(deckhouseRelease))
	}

	sort.Sort(sort.Reverse(byVersion(releases)))

	var (
		deployedReleasesIndexes []int
		outdatedReleasesIndexes []int
	)

	for i, release := range releases {
		if release.Phase == v1alpha1.PhaseDeployed {
			deployedReleasesIndexes = append(deployedReleasesIndexes, i)
		} else if release.Phase == v1alpha1.PhaseOutdated {
			outdatedReleasesIndexes = append(outdatedReleasesIndexes, i)
		}
	}

	if len(deployedReleasesIndexes) > 1 {
		// cleanup releases stacked in Deployed status
		sp := statusPatch{
			Phase: v1alpha1.PhaseOutdated,
		}
		// everything except the last Deployed release
		for i := 1; i < len(deployedReleasesIndexes); i++ {
			release := releases[i]
			input.PatchCollector.MergePatch(sp, "deckhouse.io/v1alpha1", "DeckhouseRelease", "", release.Name, object_patch.WithSubresource("/status"))
		}
	}

	// save only last 10 outdated releases
	if len(outdatedReleasesIndexes) > 10 {
		for i := 10; i < len(outdatedReleasesIndexes); i++ {
			release := releases[i]
			input.PatchCollector.Delete("deckhouse.io/v1alpha1", "DeckhouseRelease", "", release.Name, object_patch.InBackground())
		}
	}

	return nil
}
