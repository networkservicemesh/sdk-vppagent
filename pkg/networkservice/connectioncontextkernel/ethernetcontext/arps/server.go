// Copyright (c) 2020 Doc.ai and/or its affiliates.
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

// Package arps provides networkservice chain elements for setting the arp entries for kernel linux config
package arps

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"go.ligato.io/vpp-agent/v3/proto/ligato/linux"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type setKernelArpsServer struct{}

// NewServer provides a NetworkServiceServer that sets the arp entry for kernel linux config
func NewServer() networkservice.NetworkServiceServer {
	return &setKernelArpsServer{}
}

func (s *setKernelArpsServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	config := vppagent.Config(ctx)
	if mechanism := kernel.ToMechanism(request.GetConnection().GetMechanism()); mechanism != nil && len(config.LinuxConfig.Interfaces) > 0 {
		iface := config.LinuxConfig.Interfaces[len(config.LinuxConfig.Interfaces)-1]
		if request.GetConnection().GetContext().GetEthernetContext().GetDstMac() != "" &&
			request.GetConnection().GetContext().GetIpContext().GetDstIpAddr() != "" {
			config.LinuxConfig.ArpEntries = append(config.LinuxConfig.ArpEntries, &linux.ARPEntry{
				IpAddress: strings.Split(request.GetConnection().GetContext().IpContext.GetDstIpAddr(), "/")[0],
				Interface: iface.Name,
				HwAddress: request.Connection.GetContext().EthernetContext.GetDstMac(),
			})
		}
	}
	return next.Server(ctx).Request(ctx, request)
}

func (s *setKernelArpsServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	config := vppagent.Config(ctx)
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil && len(config.LinuxConfig.Interfaces) > 0 {
		iface := config.LinuxConfig.Interfaces[len(config.LinuxConfig.Interfaces)-1]
		if conn.GetContext().GetEthernetContext().GetDstMac() != "" &&
			conn.GetContext().GetIpContext().GetDstIpAddr() != "" {
			config.LinuxConfig.ArpEntries = append(config.LinuxConfig.ArpEntries, &linux.ARPEntry{
				IpAddress: strings.Split(conn.GetContext().IpContext.DstIpAddr, "/")[0],
				Interface: iface.Name,
				HwAddress: conn.GetContext().EthernetContext.GetDstMac(),
			})
		}
	}
	return next.Server(ctx).Close(ctx, conn)
}
