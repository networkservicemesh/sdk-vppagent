// Copyright (c) 2020 Doc.ai and/or its affiliates.
//
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

package macaddress

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/connection"
	"github.com/networkservicemesh/api/pkg/api/connection/mechanisms/kernel"
	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type setKernelMacClient struct{}

// NewClient provides a NetworkServiceClient that sets the mac address on a kernel interface
// It sets the Mac Address on the *kernel* side of an interface leaving the
// Client.  Generally only used by privileged Clients like those implementing
// the Cross Connect Network Service for K8s (formerly known as NSM Forwarder).
//                                         Client
//                              +---------------------------+
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           +-------------------+
//                              |                           |          macaddress.NewClient()
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              +---------------------------+
//
func NewClient() networkservice.NetworkServiceClient {
	return &setKernelMacClient{}
}

func (c *setKernelMacClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*connection.Connection, error) {
	config := vppagent.Config(ctx)
	if mechanism := kernel.ToMechanism(request.GetConnection().GetMechanism()); mechanism != nil && len(config.GetLinuxConfig().GetInterfaces()) > 0 {
		current := len(config.LinuxConfig.Interfaces) - 1
		config.LinuxConfig.Interfaces[current].PhysAddress = request.GetConnection().GetContext().GetEthernetContext().GetSrcMac()
	}
	return next.Client(ctx).Request(ctx, request, opts...)
}

func (c *setKernelMacClient) Close(ctx context.Context, conn *connection.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	config := vppagent.Config(ctx)
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil && len(config.GetLinuxConfig().GetInterfaces()) > 0 {
		current := len(config.LinuxConfig.Interfaces) - 1
		config.LinuxConfig.Interfaces[current].PhysAddress = conn.GetContext().GetEthernetContext().GetSrcMac()
	}
	return next.Client(ctx).Close(ctx, conn, opts...)
}
