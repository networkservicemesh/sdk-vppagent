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

	linuxinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linuxnamespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vppinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/common"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

func appendInterfaceConfig(ctx context.Context, conn *networkservice.Connection, name string, fileNameFromInodeNumberFunc func(string) (string, error)) (*linuxinterfaces.Interface, error) {
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil {
		conf := vppagent.Config(ctx)
		filepath, err := fileNameFromInodeNumberFunc(mechanism.GetParameters()[common.NetNSInodeKey])
		if err != nil {
			return nil, err
		}
		linuxName := name
		if len(linuxName) > kernel.LinuxIfMaxLength {
			linuxName = linuxName[:kernel.LinuxIfMaxLength]
		}
		conf.GetLinuxConfig().Interfaces = append(conf.GetLinuxConfig().Interfaces,
			&linuxinterfaces.Interface{
				Name:       name + "-veth",
				Type:       linuxinterfaces.Interface_VETH,
				Enabled:    true,
				HostIfName: linuxName,
				Link: &linuxinterfaces.Interface_Veth{
					Veth: &linuxinterfaces.VethLink{
						PeerIfName:           name,
						RxChecksumOffloading: linuxinterfaces.VethLink_CHKSM_OFFLOAD_DISABLED,
						TxChecksumOffloading: linuxinterfaces.VethLink_CHKSM_OFFLOAD_DISABLED,
					},
				},
			},
			&linuxinterfaces.Interface{
				Name:       name,
				Type:       linuxinterfaces.Interface_VETH,
				Enabled:    true,
				HostIfName: mechanism.GetInterfaceName(conn),
				Namespace: &linuxnamespace.NetNamespace{
					Type:      linuxnamespace.NetNamespace_FD,
					Reference: filepath,
				},
				Link: &linuxinterfaces.Interface_Veth{
					Veth: &linuxinterfaces.VethLink{
						PeerIfName:           name + "-veth",
						RxChecksumOffloading: linuxinterfaces.VethLink_CHKSM_OFFLOAD_DISABLED,
						TxChecksumOffloading: linuxinterfaces.VethLink_CHKSM_OFFLOAD_DISABLED,
					},
				},
			})
		conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().Interfaces, &vppinterfaces.Interface{
			Name:    name,
			Type:    vppinterfaces.Interface_AF_PACKET,
			Enabled: true,
			Link: &vppinterfaces.Interface_Afpacket{
				Afpacket: &vppinterfaces.AfpacketLink{
					HostIfName: linuxName,
				},
			},
		})
		index := len(conf.GetLinuxConfig().GetInterfaces()) - 1
		return conf.GetLinuxConfig().GetInterfaces()[index], nil
	}
	return nil, nil
}
