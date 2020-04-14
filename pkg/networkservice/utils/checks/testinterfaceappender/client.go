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

// Package testinterfaceappender provides networkservice chain elements that appends the memif interface to vppConfig.Interfaces
package testinterfaceappender

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vppinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type testInterfaceAppenderClient struct{}

// NewClient - returns a NetworkServiceClient chain elements that appends the memif interface to vppConfig.Interfaces
func NewClient() networkservice.NetworkServiceClient {
	return &testInterfaceAppenderClient{}
}

func (t *testInterfaceAppenderClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	conf := vppagent.Config(ctx)
	conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().Interfaces, &vpp.Interface{
		Name:    fmt.Sprintf("client-%s", request.GetConnection().GetId()),
		Type:    vppinterfaces.Interface_MEMIF,
		Enabled: true,
		Link: &vppinterfaces.Interface_Memif{
			Memif: &vppinterfaces.MemifLink{
				Master: false,
			},
		},
	})
	return next.Client(ctx).Request(ctx, request, opts...)
}

func (t *testInterfaceAppenderClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	conf := vppagent.Config(ctx)
	conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().Interfaces, &vpp.Interface{
		Name:    fmt.Sprintf("client-%s", conn.GetId()),
		Type:    vppinterfaces.Interface_MEMIF,
		Enabled: true,
		Link: &vppinterfaces.Interface_Memif{
			Memif: &vppinterfaces.MemifLink{
				Master: false,
			},
		},
	})
	return next.Client(ctx).Close(ctx, conn, opts...)
}
