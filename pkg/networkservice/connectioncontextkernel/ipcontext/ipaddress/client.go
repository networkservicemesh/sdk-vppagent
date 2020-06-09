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

package ipaddress

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type setIPKernelClient struct{}

// NewClient provides a NetworkServiceClient that sets the IP on a kernel interface
// It sets the IP Address on the *kernel* side of an interface leaving the
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
//                              |                           |          ipaddress.NewClient()
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              +---------------------------+
//
func NewClient() networkservice.NetworkServiceClient {
	return &setIPKernelClient{}
}

func (s *setIPKernelClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	conn, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	conf := vppagent.Config(ctx)
	if mechanism := kernel.ToMechanism(request.GetConnection().GetMechanism()); mechanism != nil && len(conf.GetLinuxConfig().GetInterfaces()) > 0 {
		index := len(conf.GetLinuxConfig().GetInterfaces()) - 1
		srcIP := conn.GetContext().GetIpContext().GetSrcIpAddr()
		if srcIP != "" {
			conf.GetLinuxConfig().GetInterfaces()[index].IpAddresses = []string{srcIP}
		}
	}
	return conn, nil
}

func (s *setIPKernelClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	e, err := next.Client(ctx).Close(ctx, conn, opts...)
	if err != nil {
		return nil, err
	}
	conf := vppagent.Config(ctx)
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil && len(conf.GetLinuxConfig().GetInterfaces()) > 0 {
		index := len(conf.GetLinuxConfig().GetInterfaces()) - 1
		srcIP := conn.GetContext().GetIpContext().GetSrcIpAddr()
		if srcIP != "" {
			conf.GetLinuxConfig().GetInterfaces()[index].IpAddresses = []string{conn.GetContext().GetIpContext().GetSrcIpAddr()}
		}
	}
	return e, err
}
