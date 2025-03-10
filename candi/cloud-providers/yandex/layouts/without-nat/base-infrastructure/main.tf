# Copyright 2021 Flant JSC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

resource "yandex_vpc_network" "kube" {
  count = local.existing_network_id != "" ? 0 : 1
  name = local.prefix
}

locals {
  network_id = local.existing_network_id != "" ? local.existing_network_id : join("", yandex_vpc_network.kube.*.id) # https://github.com/hashicorp/terraform/issues/23222#issuecomment-547462883
}

module "vpc_components" {
  source = "../../../terraform-modules/vpc-components"
  prefix = local.prefix
  network_id = local.network_id
  node_network_cidr = local.node_network_cidr
  dhcp_domain_name = local.dhcp_domain_name
  dhcp_domain_name_servers = local.dhcp_domain_name_servers

  labels = local.labels
}
