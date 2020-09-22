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

package kerneltap_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	linuxnamespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
)

const (
	netnsFileURL = "/proc/12/ns/net"
)

func checkVppAgentConfig(prefix string, request *networkservice.NetworkServiceRequest) func(*testing.T, *configurator.Config) {
	return func(t *testing.T, conf *configurator.Config) {
		require.Greater(t, len(conf.GetVppConfig().GetInterfaces()), 0)
		numInterfaces := len(conf.GetVppConfig().GetInterfaces())
		vppInterface := conf.GetVppConfig().GetInterfaces()[numInterfaces-1]
		assert.NotNil(t, vppInterface)
		assert.Equal(t, fmt.Sprintf("%s-%s", prefix, request.GetConnection().GetId()), vppInterface.GetName())
		tap := vppInterface.GetTap()
		assert.NotNil(t, tap)
		assert.Equal(t, tap.GetVersion(), uint32(2))
		linuxInterfaces := conf.GetLinuxConfig().GetInterfaces()
		assert.Greater(t, len(linuxInterfaces), 0)
		linuxInterface := linuxInterfaces[0]
		assert.NotNil(t, linuxInterface)
		assert.NotNil(t, linuxInterface.GetTap())
		assert.Equal(t, vppInterface.GetName(), linuxInterface.GetTap().GetVppTapIfName())
		assert.NotNil(t, linuxInterface.GetNamespace())
		assert.Equal(t, linuxnamespace.NetNamespace_FD, linuxInterface.GetNamespace().GetType())
		assert.Equal(t, netnsFileURL, linuxInterface.GetNamespace().GetReference())
		kmech := kernel.ToMechanism(&networkservice.Mechanism{
			Cls:  cls.LOCAL,
			Type: kernel.MECHANISM,
		})
		hostIfaceName := kmech.GetInterfaceName(request.GetConnection())
		assert.Equal(t, hostIfaceName, linuxInterface.GetHostIfName())
	}
}
