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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = Describe("Modules :: automatic-namespaces-discovery :: hooks :: discover_namespaces ::", func() {

	Context("Empty config", func() {
		f := HookExecutionConfigInit(`{}`, `{}`)

		BeforeEach(func() {
			f.KubeStateSet(`
---
apiVersion: v1
kind: Namespace
metadata:
  name: test1
---
apiVersion: v1
kind: Namespace
metadata:
  name: test2
  annotations:
    extended-monitoring.flant.com/enabled: "true"
`)
			f.BindingContexts.Set(f.GenerateBeforeHelmContext())
			f.RunHook()
		})

		It("Namespace annotations should not change", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.PatchCollector.Operations()).To(HaveLen(0))

			ns1 := f.KubernetesResource("Namespace", "", "test1")
			Expect(ns1.Field(`metadata.annotations.extended-monitoring\.flant\.com/enabled`).Exists()).To(BeFalse())

			ns2 := f.KubernetesResource("Namespace", "", "test2")
			Expect(ns2.Field(`metadata.annotations.extended-monitoring\.flant\.com/enabled`).Exists()).To(BeTrue())

		})
	})

	Context("Static matching", func() {

		f := HookExecutionConfigInit(`{"automaticNamespaceDiscovery":{"includeNames":["test1"],"excludeNames":["test2"]}}`, `{}`)

		BeforeEach(func() {
			f.KubeStateSet(`
---
apiVersion: v1
kind: Namespace
metadata:
  name: test1
---
apiVersion: v1
kind: Namespace
metadata:
  name: test2
  annotations:
    extended-monitoring.flant.com/enabled: "true"
`)
			f.BindingContexts.Set(f.GenerateBeforeHelmContext())
			f.RunHook()
		})

		It("Namespace annotations should change", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.PatchCollector.Operations()).To(HaveLen(2))

			ns1 := f.KubernetesResource("Namespace", "", "test1")
			Expect(ns1.Field(`metadata.annotations.extended-monitoring\.flant\.com/enabled`).Exists()).To(BeTrue())

			ns2 := f.KubernetesResource("Namespace", "", "test2")
			Expect(ns2.Field(`metadata.annotations.extended-monitoring\.flant\.com/enabled`).Exists()).To(BeFalse())
		})
	})

})
