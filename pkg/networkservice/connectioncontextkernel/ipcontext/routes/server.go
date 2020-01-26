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

// Package routes provides a NetworkServiceServer that sets the routes in the kernel from the connection context
package routes

import (
	"context"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ligato/vpp-agent/api/models/linux"
	linuxl3 "github.com/ligato/vpp-agent/api/models/linux/l3"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection/mechanisms/kernel"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type setKernelRoute struct{}

// NewServer creates a NetworkServiceServer that will put the routes from the connection context into
//  connection context into the kernel network namespace kernel interface being inserted iff the
//  selected mechanism for the connection is a kernel mechanism
//                                                       Endpoint
//  +- - - - - - - - - - - - - - - -+         +---------------------------+
//  |    kernel network namespace   |         |                           |
//                                            |                           |
//  |                               |         |                           |
//                                            |                           |
//  |                               |         |                           |
//                                            |                           |
//  |                               |         |                           |
//                        +--------- ---------+                           |
//  |                               |         |                           |
//                                            |                           |
//  |                               |         |                           |
//      routes.NewServer()                    |                           |
//  |                               |         |                           |
//                                            |                           |
//  |                               |         |                           |
//  +- - - - - - - - - - - - - - - -+         +---------------------------+
//
func NewServer() networkservice.NetworkServiceServer {
	return &setKernelRoute{}
}

func (s *setKernelRoute) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*connection.Connection, error) {
	s.addRoutes(ctx, request.GetConnection())
	return next.Server(ctx).Request(ctx, request)
}

func (s *setKernelRoute) Close(ctx context.Context, conn *connection.Connection) (*empty.Empty, error) {
	s.addRoutes(ctx, conn)
	return next.Server(ctx).Close(ctx, conn)
}

func (s *setKernelRoute) addRoutes(ctx context.Context, conn *connection.Connection) {
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil {
		duplicatedPrefixes := make(map[string]bool)
		for _, route := range conn.GetContext().GetIpContext().GetSrcRoutes() {
			if _, ok := duplicatedPrefixes[route.Prefix]; !ok {
				duplicatedPrefixes[route.Prefix] = true
				vppagent.Config(ctx).GetLinuxConfig().Routes = append(vppagent.Config(ctx).GetLinuxConfig().Routes, &linux.Route{
					DstNetwork:        route.Prefix,
					OutgoingInterface: vppagent.Config(ctx).GetLinuxConfig().GetInterfaces()[0].GetName(),
					Scope:             linuxl3.Route_GLOBAL,
					GwAddr:            extractCleanIPAddress(conn.GetContext().GetIpContext().GetDstIpAddr()),
				})
			}
		}
	}
}

func extractCleanIPAddress(addr string) string {
	ip, _, err := net.ParseCIDR(addr)
	if err == nil {
		return ip.String()
	}
	return addr
}
