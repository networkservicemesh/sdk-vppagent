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

package kernelvethpair

import (
	"context"

	linuxinterfaces "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	linuxnamespace "github.com/ligato/vpp-agent/api/models/linux/namespace"
	vppinterfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection/mechanisms/common"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection/mechanisms/kernel"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservicemesh/vppagent"
	"github.com/networkservicemesh/sdk-vppagent/pkg/tools/netnsinode"
)

func appendInterfaceConfig(ctx context.Context, conn *connection.Connection, name string) error {
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil {
		conf := vppagent.Config(ctx)
		filepath, err := netnsinode.LinuxNetNsFileName(mechanism.GetParameters()[common.NetNsInodeKey])
		if err != nil {
			return err
		}
		logrus.Info("Did Not Find /dev/vhost-net - using veth pairs")
		conf.GetLinuxConfig().Interfaces = append(conf.GetLinuxConfig().Interfaces,
			&linuxinterfaces.Interface{
				Name:       name + "-veth",
				Type:       linuxinterfaces.Interface_VETH,
				Enabled:    true,
				HostIfName: name + "-veth",
				Link: &linuxinterfaces.Interface_Veth{
					Veth: &linuxinterfaces.VethLink{
						PeerIfName: name,
					},
				},
			},
			&linuxinterfaces.Interface{
				Name:       name,
				Type:       linuxinterfaces.Interface_VETH,
				Enabled:    true,
				HostIfName: mechanism.GetParameters()[common.InterfaceNameKey],
				Namespace: &linuxnamespace.NetNamespace{
					Type:      linuxnamespace.NetNamespace_FD,
					Reference: filepath,
				},
				Link: &linuxinterfaces.Interface_Veth{
					Veth: &linuxinterfaces.VethLink{
						PeerIfName: name + "-veth",
					},
				},
			})
		conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().Interfaces, &vppinterfaces.Interface{
			Name:    name,
			Type:    vppinterfaces.Interface_AF_PACKET,
			Enabled: true,
			Link: &vppinterfaces.Interface_Afpacket{
				Afpacket: &vppinterfaces.AfpacketLink{
					HostIfName: name + "-veth",
				},
			},
		})
	}
	return nil
}
