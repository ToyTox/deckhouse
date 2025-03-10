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
	"errors"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/deckhouse/deckhouse/go_lib/dependency"
	"github.com/deckhouse/deckhouse/go_lib/dependency/cr"
	"github.com/deckhouse/deckhouse/go_lib/dependency/requirements"
	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = Describe("Modules :: deckhouse :: hooks :: update deckhouse image ::", func() {
	f := HookExecutionConfigInit(`{
        "global": {
          "modulesImages": {
			"registry": "my.registry.com/deckhouse"
		  }
        },
		"deckhouse": {
              "internal": {},
              "releaseChannel": "Stable",
			  "update": {
				"mode": "Auto",
				"windows": [{"from": "00:00", "to": "23:00"}]
			  }
			}
}`, `{}`)
	f.RegisterCRD("deckhouse.io", "v1alpha1", "DeckhouseRelease", false)

	dependency.TestDC.CRClient = cr.NewClientMock(GinkgoT())

	Context("Update out of window", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("deckhouse.update.windows", []byte(`[{"from": "8:00", "to": "10:00"}]`))

			f.KubeStateSet(deckhousePodYaml + deckhouseReleases)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})
		It("Should keep deckhouse deployment", func() {
			Expect(f).To(ExecuteSuccessfully())
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.25.0"))
		})
	})

	Context("No update windows configured", func() {
		BeforeEach(func() {
			f.ValuesDelete("deckhouse.windows")

			f.KubeStateSet(deckhousePodYaml + deckhouseReleases)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})
		It("Should upgrade deckhouse deployment", func() {
			Expect(f).To(ExecuteSuccessfully())
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.26.0"))
		})
	})

	Context("Update out of day window", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("deckhouse.update.windows", []byte(`[{"from": "8:00", "to": "23:00", "days": ["Mon", "Tue"]}]`))

			f.KubeStateSet(deckhousePodYaml + deckhouseReleases)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})
		It("Should keep deckhouse deployment", func() {
			Expect(f).To(ExecuteSuccessfully())
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.25.0"))
		})
	})

	Context("Update in day window", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("deckhouse.update.windows", []byte(`[{"from": "8:00", "to": "23:00", "days": ["Fri", "Sun"]}]`))

			f.KubeStateSet(deckhousePodYaml + deckhouseReleases)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})
		It("Should upgrade deckhouse deployment", func() {
			Expect(f).To(ExecuteSuccessfully())
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.26.0"))
		})
	})

	Context("Patch out of update window", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("deckhouse.update.windows", []byte(`[{"from": "8:00", "to": "8:01"}]`))

			f.KubeStateSet(deckhousePodYaml + deckhouseReleases + deckhousePatchRelease)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})
		It("Should upgrade deckhouse deployment to patch version", func() {
			Expect(f).To(ExecuteSuccessfully())
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.25.1"))
			patchRelease := f.KubernetesGlobalResource("DeckhouseRelease", "v1-25-1")
			Expect(patchRelease.Field("status.approved").Bool()).To(Equal(true))
		})
	})

	Context("Deckhouse previous release is not ready", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("deckhouse.update.windows", []byte(`[{"from": "00:00", "to": "23:59"}]`))

			f.KubeStateSet(deckhouseDeployment + deckhouseNotReadyPod + deckhouseReleases)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})
		It("Should not upgrade deckhouse version", func() {
			Expect(f).To(ExecuteSuccessfully())
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.25.0"))
		})
	})

	Context("Manual approval mode is set", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("deckhouse.update.mode", []byte(`"Manual"`))
			f.ValuesDelete("deckhouse.update.windows")

			f.KubeStateSet(deckhouseDeployment + deckhouseReadyPod + deckhouseReleases)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})
		It("Should not upgrade deckhouse version", func() {
			Expect(f).To(ExecuteSuccessfully())
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.25.0"))
			Expect(f.MetricsCollector.CollectedMetrics()).To(HaveLen(2))
		})

		Context("After setting manual approve", func() {
			BeforeEach(func() {
				f.KubeStateSet("")
				cc := f.KubeStateSet(deckhouseDeployment + deckhouseReadyPod + manualApprovedReleases)
				f.BindingContexts.Set(cc)
				f.RunHook()
			})
			It("Must upgrade deckhouse", func() {
				Expect(f).To(ExecuteSuccessfully())
				dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
				Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(Equal("my.registry.com/deckhouse:v1.26.0"))
			})
		})

		Context("Auto deploy Patch release in Manual mode", func() {
			BeforeEach(func() {
				f.KubeStateSet("")
				cc := f.KubeStateSet(deckhouseDeployment + deckhouseReadyPod + deckhouseReleases + deckhousePatchRelease)
				f.BindingContexts.Set(cc)
				f.RunHook()
			})
			It("Must upgrade deckhouse", func() {
				Expect(f).To(ExecuteSuccessfully())
				dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
				Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(Equal("my.registry.com/deckhouse:v1.25.1"))
			})
		})
	})

	Context("DEV: No new deckhouse image", func() {
		BeforeEach(func() {
			dependency.TestDC.CRClient.DigestMock.Set(func(tag string) (s1 string, err error) {
				return "sha256:d57f01a88e54f863ff5365c989cb4e2654398fa274d46389e0af749090b862d1", nil
			})
			f.KubeStateSet(deckhousePodYaml + deckhouseReleases)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})
		It("Should keep deckhouse pod", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.KubernetesResource("Pod", "d8-system", "deckhouse-6f46df5bd7-nk4j7").Exists()).To(BeTrue())
		})
	})

	Context("DEV: Have new deckhouse image", func() {
		BeforeEach(func() {
			dependency.TestDC.CRClient.DigestMock.Set(func(tag string) (s1 string, err error) {
				return "sha256:123456", nil
			})
			f.KubeStateSet(deckhousePodYaml)
			f.ValuesDelete("deckhouse.releaseChannel")
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})
		It("Should remove deckhouse pod", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.KubernetesResource("Pod", "d8-system", "deckhouse-6f46df5bd7-nk4j7").Exists()).To(BeFalse())
		})
	})

	Context("Manual mode", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("deckhouse.update.mode", []byte(`"Manual"`))
			f.ValuesDelete("deckhouse.update.windows")

			f.KubeStateSet(newReleaseState)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})

		It("Should keep deckhouse deployment", func() {
			Expect(f).To(ExecuteSuccessfully())
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.26.2"))
		})

		Context("Second run of the hook in a Manual mode should not change state", func() {
			BeforeEach(func() {
				f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
				f.RunHook()
			})

			It("Should keep deckhouse deployment", func() {
				Expect(f).To(ExecuteSuccessfully())
				dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
				Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.26.2"))
			})
		})
	})

	Context("Single First Release", func() {
		BeforeEach(func() {
			f.KubeStateSet(deckhousePodYaml + deckhousePatchRelease)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})

		It("Should update deckhouse deployment", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.KubernetesGlobalResource("DeckhouseRelease", "v1-25-1").Field("status.phase").String()).To(Equal("Deployed"))
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.25.1"))
		})
	})

	Context("Postponed release", func() {
		BeforeEach(func() {
			f.KubeStateSet(deckhousePodYaml + postponedRelease)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})

		It("Should keep deckhouse deployment", func() {
			Expect(f).To(ExecuteSuccessfully())
			r1250 := f.KubernetesGlobalResource("DeckhouseRelease", "v1-25-0")
			r1251 := f.KubernetesGlobalResource("DeckhouseRelease", "v1-25-1")
			Expect(r1250.Field("status.phase").String()).To(Equal("Deployed"))
			Expect(r1251.Field("status.phase").String()).To(Equal("Pending"))
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.25.0"))
		})

		Context("Release applyAfter time passed", func() {
			BeforeEach(func() {
				f.KubeStateSet(deckhousePodYaml + strings.Replace(postponedRelease, "2222-11-11T23:23:23Z", "2001-11-11T23:23:23Z", 1))

				f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
				f.RunHook()
			})

			It("Should upgrade deckhouse deployment", func() {
				Expect(f).To(ExecuteSuccessfully())
				r1250 := f.KubernetesGlobalResource("DeckhouseRelease", "v1-25-0")
				r1251 := f.KubernetesGlobalResource("DeckhouseRelease", "v1-25-1")
				Expect(r1250.Field("status.phase").String()).To(Equal("Outdated"))
				Expect(r1251.Field("status.phase").String()).To(Equal("Deployed"))
				dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
				Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.25.1"))
			})
		})
	})

	Context("Suspend release", func() {
		BeforeEach(func() {
			f.KubeStateSet(deckhousePodYaml + withSuspendedRelease)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})

		It("Should update deckhouse deployment", func() {
			Expect(f).To(ExecuteSuccessfully())
			r1250 := f.KubernetesGlobalResource("DeckhouseRelease", "v1-25-0")
			r1251 := f.KubernetesGlobalResource("DeckhouseRelease", "v1-25-1")
			r1252 := f.KubernetesGlobalResource("DeckhouseRelease", "v1-25-2")
			Expect(r1250.Field("status.phase").String()).To(Equal("Outdated"))
			Expect(r1251.Field("status.phase").String()).To(Equal("Suspended"))
			Expect(r1252.Field("status.phase").String()).To(Equal("Deployed"))
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.25.2"))
		})
	})

	Context("Release with not met requirements", func() {
		BeforeEach(func() {
			requirements.Register("k8s", func(requirementValue string, getter requirements.ValueGetter) (bool, error) {
				v := getter.Get("global.discovery.kubernetesVersion").String()
				if v != requirementValue {
					return false, errors.New("min k8s version failed")
				}

				return true, nil
			})
			f.ValuesSet("global.discovery.kubernetesVersion", "1.16.0")
			f.KubeStateSet(deckhousePodYaml + releaseWithRequirements)
			f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
			f.RunHook()
		})

		It("Should not update deckhouse deployment", func() {
			Expect(f).To(ExecuteSuccessfully())
			r130 := f.KubernetesGlobalResource("DeckhouseRelease", "v1-30-0")
			Expect(r130.Field("status.phase").String()).To(Equal("Pending"))
			Expect(r130.Field("status.message").String()).To(Equal(`"k8s" requirement for DeckhouseRelease "1.30.0" not met: min k8s version failed`))
			Expect(f.MetricsCollector.CollectedMetrics()[1].Name).To(Equal("d8_release_blocked"))
			Expect(*f.MetricsCollector.CollectedMetrics()[1].Value).To(Equal(float64(1)))
			dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
			Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.25.0"))
		})

		Context("Release requirements passed", func() {
			BeforeEach(func() {
				f.ValuesSet("global.discovery.kubernetesVersion", "1.19.0")
				f.BindingContexts.Set(f.GenerateScheduleContext("*/15 * * * * *"))
				f.RunHook()
			})

			It("Should update deckhouse deployment", func() {
				Expect(f).To(ExecuteSuccessfully())
				r130 := f.KubernetesGlobalResource("DeckhouseRelease", "v1-30-0")
				Expect(r130.Field("status.phase").String()).To(Equal("Deployed"))
				Expect(r130.Field("status.message").String()).To(Equal(``))
				dep := f.KubernetesResource("Deployment", "d8-system", "deckhouse")
				Expect(dep.Field("spec.template.spec.containers").Array()[0].Get("image").String()).To(BeEquivalentTo("my.registry.com/deckhouse:v1.30.0"))
			})
		})
	})

})

var (
	deckhouseReadyPod = `
---
apiVersion: v1
kind: Pod
metadata:
  name: deckhouse-6f46df5bd7-nk4j7
  namespace: d8-system
  labels:
    app: deckhouse
spec:
  containers:
    - name: deckhouse
      image: dev-registry.deckhouse.io/sys/deckhouse-oss/dev:test-me
status:
  containerStatuses:
    - containerID: containerd://9990d3eccb8657d0bfe755672308831b6d0fab7f3aac553487c60bf0f076b2e3
      imageID: dev-registry.deckhouse.io/sys/deckhouse-oss/dev@sha256:d57f01a88e54f863ff5365c989cb4e2654398fa274d46389e0af749090b862d1
      ready: true
`
	deckhouseNotReadyPod = `
---
apiVersion: v1
kind: Pod
metadata:
  name: deckhouse-6f46df5bd7-nk4j7
  namespace: d8-system
  labels:
    app: deckhouse
spec:
  containers:
    - name: deckhouse
      image: dev-registry.deckhouse.io/sys/deckhouse-oss/dev:test-me
status:
  containerStatuses:
    - containerID: containerd://9990d3eccb8657d0bfe755672308831b6d0fab7f3aac553487c60bf0f076b2e3
      imageID: dev-registry.deckhouse.io/sys/deckhouse-oss/dev@sha256:d57f01a88e54f863ff5365c989cb4e2654398fa274d46389e0af749090b862d1
      ready: false
`

	deckhouseDeployment = `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deckhouse
  namespace: d8-system
spec:
  template:
    spec:
      containers:
        - name: deckhouse
          image: my.registry.com/deckhouse:v1.25.0
`

	deckhousePodYaml = deckhouseReadyPod + deckhouseDeployment

	deckhouseReleases = `
---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-25-0
spec:
  version: "v1.25.0"
status:
  phase: Deployed
---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-26-0
spec:
  version: "v1.26.0"
`

	manualApprovedReleases = `
---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-25-0
spec:
  version: "v1.25.0"
status:
  phase: Deployed
---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-26-0
spec:
  version: "v1.26.0"
approved: true
`

	deckhousePatchRelease = `

---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-25-1
spec:
  version: "v1.25.1"
`

	newReleaseState = `
---
apiVersion: deckhouse.io/v1alpha1
approved: false
kind: DeckhouseRelease
metadata:
  name: v1-26-2
spec:
  version: v1.26.2
status:
  approved: true
  phase: Deployed
  transitionTime: "2021-12-08T08:34:01.292180321Z"
---
apiVersion: deckhouse.io/v1alpha1
approved: false
kind: DeckhouseRelease
metadata:
  name: v1-27-0
spec:
  version: v1.27.0
---
apiVersion: v1
kind: Pod
metadata:
  name: deckhouse-6f46df5bd7-nk4j7
  namespace: d8-system
  labels:
    app: deckhouse
spec:
  containers:
    - name: deckhouse
      image: dev-registry.deckhouse.io/sys/deckhouse-oss/dev:1-26-2
status:
  containerStatuses:
    - containerID: containerd://9990d3eccb8657d0bfe755672308831b6d0fab7f3aac553487c60bf0f076b2e3
      imageID: dev-registry.deckhouse.io/sys/deckhouse-oss/dev@sha256:d57f01a88e54f863ff5365c989cb4e2654398fa274d46389e0af749090b862d1
      ready: true
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deckhouse
  namespace: d8-system
spec:
  template:
    spec:
      containers:
        - name: deckhouse
          image: my.registry.com/deckhouse:v1.26.2
`

	withSuspendedRelease = `
---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-25-0
spec:
  version: "v1.25.0"
status:
  phase: Deployed
---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-25-1
  annotations:
    release.deckhouse.io/suspended: "true"
spec:
  version: "v1.25.1"
status:
  phase: Pending
---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-25-2
spec:
  version: "v1.25.2"
status:
  phase: Pending
`

	postponedRelease = `
---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-25-0
spec:
  version: "v1.25.0"
status:
  phase: Deployed
---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-25-1
spec:
  version: "v1.25.1"
  applyAfter: "2222-11-11T23:23:23Z"
status:
  phase: Pending
`

	releaseWithRequirements = `
---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-29-0
spec:
  version: "v1.29.0"
status:
  phase: Deployed
---
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  name: v1-30-0
spec:
  version: "v1.30.0"
  requirements:
    k8s: "1.19.0"
status:
  phase: Pending
`
)
