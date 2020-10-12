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
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type setVppRoutesClient struct{}

// NewClient creates a NetworkServiceClient chain element to set routes in vpp
//                                         Client
//                              +---------------------------+
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |      routes.NewClient()   +-------------------+
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              +---------------------------+
//
func NewClient() networkservice.NetworkServiceClient {
	return &setVppRoutesClient{}
}

func (s *setVppRoutesClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	conn, err := next.Client(ctx).Request(ctx, request)
	if err != nil {
		return nil, err
	}
	s.addRoutes(ctx, conn)
	return conn, nil
}

func (s *setVppRoutesClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	rv, err := next.Client(ctx).Close(ctx, conn)
	if err != nil {
		return nil, err
	}
	s.addRoutes(ctx, conn)
	return rv, nil
}

func (s *setVppRoutesClient) addRoutes(ctx context.Context, conn *networkservice.Connection) {
	// If we aren't plugging in an interface... nothing to do here
	if len(vppagent.Config(ctx).GetVppConfig().GetInterfaces()) == 0 {
		return
	}
	// If the interface is nil, nothing to do
	iface := vppagent.Config(ctx).GetVppConfig().GetInterfaces()[0]
	if iface == nil {
		return
	}
	// Extract the dstIP and DstNet
	dstIP, dstNet, err := net.ParseCIDR(conn.GetContext().GetIpContext().GetDstIpAddr())
	if err != nil {
		return
	}
	_, srcNet, err := net.ParseCIDR(conn.GetContext().GetIpContext().GetSrcIpAddr())
	if err != nil {
		return
	}

	// Loop over any explicit routes returned with the ConnectionContext and add them
	duplicatedPrefixes := make(map[string]bool)
	for _, route := range conn.GetContext().GetIpContext().GetSrcRoutes() {
		if _, ok := duplicatedPrefixes[route.Prefix]; !ok {
			duplicatedPrefixes[route.Prefix] = true
			vppagent.Config(ctx).GetVppConfig().Routes = append(vppagent.Config(ctx).GetVppConfig().Routes, &vpp.Route{
				DstNetwork:        route.Prefix,
				OutgoingInterface: iface.GetName(),
				NextHopAddr:       dstIP.String(),
			})
		}
	}
	// If srcNet contains dstIP then dstIP is reachable and we are done
	if _, ok := duplicatedPrefixes[dstNet.String()]; ok || srcNet.Contains(dstIP) {
		return
	}

	// Otherwise add a route to dstNet. using dstIP a nextHop
	if dstIP.IsGlobalUnicast() {
		vppagent.Config(ctx).GetVppConfig().Routes = append(vppagent.Config(ctx).GetVppConfig().Routes, &vpp.Route{
			DstNetwork:        dstNet.String(),
			OutgoingInterface: iface.GetName(),
			VrfId:             iface.Vrf,
		})
	}
}
