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

// Package ipaddress provides networkservice chain elements to set the ip address on vpp interfaces
package ipaddress

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type setVppIPClient struct{}

// NewClient creates a NetworkServiceClient chain element to set the ip address on a vpp interface
// It sets the IP Address on the *vpp* side of an interface leaving the
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
//                              |      ipaddress.NewClient()+-------------------+
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
	return &setVppIPClient{}
}

func (s *setVppIPClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	conn, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	conf := vppagent.Config(ctx)
	if index := len(conf.GetVppConfig().GetInterfaces()) - 1; index >= 0 && conn.GetContext().GetIpContext().GetSrcIpAddr() != "" {
		conf.GetVppConfig().GetInterfaces()[index].IpAddresses = []string{conn.GetContext().GetIpContext().GetSrcIpAddr()}
	}
	return conn, nil
}

func (s *setVppIPClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	e, err := next.Client(ctx).Close(ctx, conn, opts...)
	conf := vppagent.Config(ctx)
	if index := len(conf.GetVppConfig().GetInterfaces()) - 1; index >= 0 && conn.GetContext().GetIpContext().GetSrcIpAddr() != "" {
		conf.GetVppConfig().GetInterfaces()[index].IpAddresses = []string{conn.GetContext().GetIpContext().GetSrcIpAddr()}
	}
	return e, err
}
