// Copyright (c) 2020 Cisco and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vppagent

const (
	vppAgentGoVPPConfFilename = vppAgentConfDir + "govpp.conf"
	vppAgentGoVPPConfContents = `# Path to a Unix-domain socket through which configuration requests are sent to VPP.
# Used if connect-via-shm is set to false and env. variable GOVPPMUX_NOSOCK is not defined.
# Default is "/run/vpp-api.sock".
binapi-socket-path: %[1]s/var/run/vpp/api.sock

# Socket path for reading VPP status, default is "/run/vpp/stats.sock"
stats-socket-path: %[1]s/var/run/vpp/stats.sock`
)
