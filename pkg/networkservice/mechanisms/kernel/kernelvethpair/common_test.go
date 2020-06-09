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

package kernelvethpair_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	linuxinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linuxnamespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
)

func fileNameFromInodeNumberFunc(inodeNum string) (string, error) {
	return "/proc/12/ns/net", nil
}

func checkVppAgentConfig(prefix string, request *networkservice.NetworkServiceRequest) func(*testing.T, *configurator.Config) {
	return func(t *testing.T, conf *configurator.Config) {
		var veths []*linuxinterfaces.VethLink

		// Basic Linux Interface Checks
		linuxInterfaces := conf.GetLinuxConfig().GetInterfaces()
		linuxInterfaces = linuxInterfaces[len(linuxInterfaces)-2:]
		require.Greater(t, len(linuxInterfaces), 1)
		for i, linuxInterface := range linuxInterfaces {
			assert.NotNil(t, linuxInterface)
			assert.LessOrEqual(t, len(linuxInterface.GetHostIfName()), kernel.LinuxIfMaxLength)
			veths = append(veths, linuxInterface.GetVeth())
			assert.NotNil(t, veths[i])
		}

		// Check Pod side interface name
		kmech := kernel.ToMechanism(&networkservice.Mechanism{
			Cls:  cls.LOCAL,
			Type: kernel.MECHANISM,
		})
		hostIfaceName := kmech.GetInterfaceName(request.GetConnection())
		assert.Equal(t, hostIfaceName, linuxInterfaces[1].GetHostIfName())

		// Check Pod side netns
		assert.NotNil(t, linuxInterfaces[1].GetNamespace())
		assert.Equal(t, linuxnamespace.NetNamespace_FD, linuxInterfaces[1].GetNamespace().GetType())
		filepath, _ := fileNameFromInodeNumberFunc("")
		assert.Equal(t, filepath, linuxInterfaces[1].GetNamespace().GetReference())

		// Check vethpair peers are correct
		assert.Equal(t, linuxInterfaces[0].GetName(), veths[1].GetPeerIfName())
		assert.Equal(t, linuxInterfaces[1].GetName(), veths[0].GetPeerIfName())

		// Check VPP Interface
		require.Greater(t, len(conf.GetVppConfig().GetInterfaces()), 0)
		numInterfaces := len(conf.GetVppConfig().GetInterfaces())
		vppInterface := conf.GetVppConfig().GetInterfaces()[numInterfaces-1]
		assert.NotNil(t, vppInterface)
		assert.Equal(t, fmt.Sprintf("%s-%s", prefix, request.GetConnection().GetId()), vppInterface.GetName())
		afpacketInterface := vppInterface.GetAfpacket()
		assert.NotNil(t, afpacketInterface)
		assert.Equal(t, afpacketInterface.GetHostIfName(), linuxInterfaces[0].GetHostIfName())
	}
}
