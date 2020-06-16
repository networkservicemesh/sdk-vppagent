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

package routes

import (
	"context"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"go.ligato.io/vpp-agent/v3/proto/ligato/linux"
	linuxl3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type setKernelRouteClient struct{}

// NewClient creates a NetworkServiceClient that will put the routes from the connection context into
//  the kernel network namespace kernel interface being inserted iff the
//  selected mechanism for the connection is a kernel mechanism
//             Client
//  +- - - - - - - - - - - - - - - -+         +---------------------------+
//  |                               |         |  kernel network namespace |
//                                            |                           |
//  |                               |         |                           |
//                                            |                           |
//  |                               |         |                           |
//                                            |                           |
//  |                               |         |                           |
//                                  +--------- ---------+                 |
//  |                               |         |                           |
//                                            |                           |
//  |                               |         |                           |
//                                            |      routes.Client()      |
//  |                               |         |                           |
//                                            |                           |
//  |                               |         |                           |
//  +- - - - - - - - - - - - - - - -+         +---------------------------+
//
func NewClient() networkservice.NetworkServiceClient {
	return &setKernelRouteClient{}
}

func (s *setKernelRouteClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	rv, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	s.addRoutes(ctx, rv)
	return rv, err
}

func (s *setKernelRouteClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	rv, err := next.Client(ctx).Close(ctx, conn, opts...)
	if err != nil {
		return nil, err
	}
	s.addRoutes(ctx, conn)
	return rv, err
}

func (s *setKernelRouteClient) addRoutes(ctx context.Context, conn *networkservice.Connection) {
	if conn.GetContext().GetIpContext().GetSrcIpAddr() == "" {
		return
	}
	srcIP, srcNet, err := net.ParseCIDR(conn.GetContext().GetIpContext().GetSrcIpAddr())
	if err != nil {
		return
	}
	conf := vppagent.Config(ctx)
	index := len(conf.GetLinuxConfig().GetInterfaces()) - 1
	if index >= 0 && srcIP.IsGlobalUnicast() {
		iface := conf.GetLinuxConfig().GetInterfaces()[index]
		vppagent.Config(ctx).GetLinuxConfig().Routes = append(vppagent.Config(ctx).GetLinuxConfig().Routes, &linux.Route{
			DstNetwork:        srcNet.String(),
			OutgoingInterface: iface.GetName(),
			Scope:             linuxl3.Route_LINK,
		})
	}
}
