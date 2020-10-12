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

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type setVppRoutesServer struct{}

// NewServer creates a NetworkServiceServer chain element to set the ip address on a vpp interface
// It sets the IP Address on the *vpp* side of an interface plugged into the
// Endpoint.
//                                         Endpoint
//                              +---------------------------+
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//          +-------------------+ ipaddress.NewServer()     |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              +---------------------------+
//
func NewServer() networkservice.NetworkServiceServer {
	return &setVppRoutesServer{}
}

func (s *setVppRoutesServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	s.addRoutes(ctx, request.GetConnection())
	return next.Server(ctx).Request(ctx, request)
}

func (s *setVppRoutesServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	s.addRoutes(ctx, conn)
	return next.Server(ctx).Close(ctx, conn)
}

func (s *setVppRoutesServer) addRoutes(ctx context.Context, conn *networkservice.Connection) {
	if conn.GetContext().GetIpContext().GetSrcIpAddr() == "" {
		return
	}
	srcIP, srcNet, err := net.ParseCIDR(conn.GetContext().GetIpContext().GetSrcIpAddr())
	if err != nil {
		return
	}
	conf := vppagent.Config(ctx)
	index := len(conf.GetVppConfig().GetInterfaces()) - 1
	if index >= 0 && srcIP.IsGlobalUnicast() {
		iface := conf.GetVppConfig().GetInterfaces()[index]
		vppagent.Config(ctx).GetVppConfig().Routes = append(vppagent.Config(ctx).GetVppConfig().Routes, &vpp.Route{
			DstNetwork:        srcNet.String(),
			OutgoingInterface: iface.GetName(),
			VrfId:             iface.Vrf,
		})
	}
}
