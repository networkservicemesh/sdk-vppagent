// Copyright (c) 2020 Cisco Systems, Inc.
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

package kernel

import (
	"os"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservicemesh/mechanisms/kernel/kerneltap"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservicemesh/mechanisms/kernel/kernelvethpair"
)

const (
	vnetFilename = "/dev/vhost-net"
)

// NewClient return a NetworkServiceClient chain element that correctly handles the kernel Mechanism
func NewClient() networkservice.NetworkServiceClient {
	if _, err := os.Stat(vnetFilename); err == nil {
		return kerneltap.NewClient()
	}
	return kernelvethpair.NewClient()
}
