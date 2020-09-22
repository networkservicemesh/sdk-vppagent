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
	"fmt"
	"net/url"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	linuxinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linuxnamespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vppinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

const (
	fileScheme = "file"
)

func appendInterfaceConfig(ctx context.Context, conn *networkservice.Connection, prefix string) error {
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil {
		netNSURLStr := mechanism.GetNetNSURL()
		netNSURL, err := url.Parse(netNSURLStr)
		if err != nil {
			return err
		}
		if netNSURL.Scheme != fileScheme {
			return errors.Errorf("kernel.ToMechanism(conn.GetMechanism()).GetNetNSURL() must be of scheme %q: %q", fileScheme, netNSURL)
		}
		vppagentConfigTemplate(vppagent.Config(ctx), fmt.Sprintf("%s-%s", prefix, conn.GetId()), kernel.ToMechanism(conn.GetMechanism()).GetInterfaceName(conn), netNSURL.Path)
	}
	return nil
}

func vppagentConfigTemplate(conf *configurator.Config, name, ifaceName, netnsFilename string) {
	conf.GetLinuxConfig().Interfaces = append(conf.GetLinuxConfig().Interfaces,
		&linuxinterfaces.Interface{
			Name:       name + "-veth",
			Type:       linuxinterfaces.Interface_VETH,
			Enabled:    true,
			HostIfName: linuxIfaceName(name),
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
			HostIfName: linuxIfaceName(ifaceName),
			Namespace: &linuxnamespace.NetNamespace{
				Type:      linuxnamespace.NetNamespace_FD,
				Reference: netnsFilename,
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
				HostIfName: linuxIfaceName(name),
			},
		},
	})
}

func linuxIfaceName(ifaceName string) string {
	if len(ifaceName) <= kernel.LinuxIfMaxLength {
		return ifaceName
	}
	return ifaceName[:kernel.LinuxIfMaxLength]
}
