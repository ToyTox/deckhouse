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

bb-is-ubuntu-version?() {
  local UBUNTU_VERSION=$1
  source /etc/os-release
  if [ "${VERSION_ID}" == "${UBUNTU_VERSION}" ] ; then
    return 0
  else
    return 1
  fi
}

bb-is-centos-version?() {
  local CENTOS_VERSION=$1
  source /etc/os-release
  if [[ "${VERSION_ID}" =~ ${CENTOS_VERSION}.* ]] ; then
    return 0
  else
    return 1
  fi
}

bb-is-debian-version?() {
  local DEBIAN_VERSION=$1
  source /etc/os-release
  if [ "${VERSION_ID}" == "${DEBIAN_VERSION}" ] ; then
    return 0
  else
    return 1
  fi
}

bb-is-astra-version?() {
  local ASTRA_VERSION_REGEX=$1
  source /etc/os-release
  if [ "${ID}" != "astra" ] ; then
    return 1
  fi
  if [[ "${VERSION_ID}" =~ ${ASTRA_VERSION_REGEX} ]] ; then
    return 0
  else
    return 1
  fi
}
