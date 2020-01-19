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

// Package setmacvpp provides networkservice chain elements for setting the mac address on vpp interfaces
package setmacvpp

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type setMacVppClient struct{}

// NewClient creates a NetworkServiceClient chain element to set the mac address on a vpp interface
// It sets the Mac Address on the *vpp* side of an interface leaving the
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
//                              |      setmacvpp.NewClient()+-------------------+
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
	return &setMacVppClient{}
}

func (s *setMacVppClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*connection.Connection, error) {
	conf := vppagent.Config(ctx)
	conn, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	index := len(conf.GetVppConfig().GetInterfaces()) - 1
	conf.GetVppConfig().GetInterfaces()[index+1].PhysAddress = conn.GetContext().GetEthernetContext().GetSrcMac()
	return conn, nil
}

func (s *setMacVppClient) Close(ctx context.Context, conn *connection.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	conf := vppagent.Config(ctx)
	conf.GetVppConfig().GetInterfaces()[len(conf.GetVppConfig().GetInterfaces())-1].PhysAddress = conn.GetContext().GetEthernetContext().GetSrcMac()
	return next.Client(ctx).Close(ctx, conn, opts...)
}
