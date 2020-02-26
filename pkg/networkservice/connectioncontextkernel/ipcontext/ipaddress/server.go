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

// Package ipaddress provides networkservice chain elements that support setting ip addresses on kernel interfaces
package ipaddress

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type setIPKernelServer struct{}

// NewServer provides a NetworkServiceServer that sets the IP on a kernel interface
// It sets the IP Address on the *kernel* side of an interface plugged into the
// Endpoint.  Generally only used by privileged Endpoints like those implementing
// the Cross Connect Network Service for K8s (formerly known as NSM Forwarder).
//                                         Endpoint
//                              +---------------------------+
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//          +-------------------+                           |
//  ipaddress.NewServer()       |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              +---------------------------+
//
func NewServer() networkservice.NetworkServiceServer {
	return &setIPKernelServer{}
}

func (s *setIPKernelServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	conf := vppagent.Config(ctx)
	if mechanism := kernel.ToMechanism(request.GetConnection().GetMechanism()); mechanism != nil && len(conf.GetLinuxConfig().GetInterfaces()) > 0 {
		index := len(conf.GetLinuxConfig().GetInterfaces()) - 1
		dstIP := request.GetConnection().GetContext().GetIpContext().GetDstIpAddr()
		conf.GetLinuxConfig().GetInterfaces()[index].IpAddresses = []string{dstIP}
	}
	return next.Client(ctx).Request(ctx, request)
}

func (s *setIPKernelServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	conf := vppagent.Config(ctx)
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil && len(conf.GetLinuxConfig().GetInterfaces()) > 0 {
		index := len(conf.GetLinuxConfig().GetInterfaces()) - 1
		dstIP := conn.GetContext().GetIpContext().GetDstIpAddr()
		conf.GetLinuxConfig().GetInterfaces()[index].IpAddresses = []string{dstIP}
	}
	return next.Client(ctx).Close(ctx, conn)
}
