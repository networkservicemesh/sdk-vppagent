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

// Package kerneltap provides networkservice chain elements that support the kernel Mechanism via tapv2
package kerneltap

import (
	"context"
	"net/url"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"

	"go.ligato.io/vpp-agent/v3/proto/ligato/linux"
	linuxinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linuxnamespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vppinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

const (
	fileScheme = "file"
)

func appendInterfaceConfig(ctx context.Context, conn *networkservice.Connection, name string) error {
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil {
		netNSURLStr := mechanism.GetNetNSURL()
		netNSURL, err := url.Parse(netNSURLStr)
		if err != nil {
			return err
		}
		if netNSURL.Scheme != fileScheme {
			return errors.Errorf("kernel.ToMechanism(conn.GetMechanism()).GetNetNSURL() must be of scheme %q: %q", fileScheme, netNSURL)
		}
		vppagentConfigTemplate(vppagent.Config(ctx), name, kernel.ToMechanism(conn.GetMechanism()).GetInterfaceName(conn), netNSURL.Path)
	}
	return nil
}

func vppagentConfigTemplate(conf *configurator.Config, name, ifaceName, netnsFilename string) {
	// We append an Interfaces.  Interfaces creates the vpp side of an interface.
	//   In this case, a Tapv2 interface that has one side in vpp, and the other
	//   as a Linux kernel interface
	conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().GetInterfaces(), &vppinterfaces.Interface{
		Name:    name,
		Type:    vppinterfaces.Interface_TAP,
		Enabled: true,
		Link: &vppinterfaces.Interface_Tap{
			Tap: &vppinterfaces.TapLink{
				Version: 2,
			},
		},
	})
	// We apply configuration to LinuxInterfaces
	// Important details:
	//    - LinuxInterfaces.HostIfName - must be no longer than 15 chars (linux limitation)
	conf.GetLinuxConfig().Interfaces = append(conf.GetLinuxConfig().Interfaces, &linux.Interface{
		Name:       name,
		Type:       linuxinterfaces.Interface_TAP_TO_VPP,
		Enabled:    true,
		HostIfName: linuxIfaceName(ifaceName),
		Namespace: &linuxnamespace.NetNamespace{
			Type:      linuxnamespace.NetNamespace_FD,
			Reference: netnsFilename,
		},
		Link: &linuxinterfaces.Interface_Tap{
			Tap: &linuxinterfaces.TapLink{
				VppTapIfName: name,
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
