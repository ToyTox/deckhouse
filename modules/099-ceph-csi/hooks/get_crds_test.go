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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = Describe("Modules :: cephCsi :: hooks :: get_crds ::", func() {
	f := HookExecutionConfigInit(`{"cephCsi":{"internal":{}}}`, ``)
	f.RegisterCRD("deckhouse.io", "v1alpha1", "CephCSIRBD", false)

	cr := `
kind: CephCSIRBD
apiVersion: deckhouse.io/v1alpha1
metadata:
  name: test
spec:
  clusterID: 5d6488af-d131-4721-b963-c2356fbbb6fb
  logLevel: 5
  monitors:
  - v2:192.168.199.210:3300/0,v1:192.168.199.210:6789/0
  storageClasses:
  - name: csi-rbd
    allowVolumeExpansion: true
    defaultFSType: ext4
    mountOptions:
    - discard
    pool: kubernetes-dev
    reclaimPolicy: Retain
  userID: kubernetes
  userKey: AQAoNqlhVc/WMRAATxrJPc+BykLMf54gSsi7yA==
`
	Context("Cluster with cr", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(cr))
			f.RunHook()
		})
		It("Value should not change", func() {
			Expect(f).To(ExecuteSuccessfully())
			fmt.Println("EEE: ", f.ValuesGet("cephCsi").String())
		})
	})

})
