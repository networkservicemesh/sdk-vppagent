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

package testinterfaceappender

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vppinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type testInterfaceAppenderServer struct{}

// NewServer - returns a NetworkServiceServer chain elements that appends the memif interface to vppConfig.Interfaces
func NewServer() networkservice.NetworkServiceServer {
	return &testInterfaceAppenderServer{}
}

func (t *testInterfaceAppenderServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	conf := vppagent.Config(ctx)
	conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().Interfaces, &vpp.Interface{
		Name:    fmt.Sprintf("server-%s", request.GetConnection().GetId()),
		Type:    vppinterfaces.Interface_MEMIF,
		Enabled: true,
		Link: &vppinterfaces.Interface_Memif{
			Memif: &vppinterfaces.MemifLink{
				Master: false,
			},
		},
	})
	return next.Server(ctx).Request(ctx, request)
}

func (t *testInterfaceAppenderServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	conf := vppagent.Config(ctx)
	conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().Interfaces, &vpp.Interface{
		Name:    fmt.Sprintf("server-%s", conn.GetId()),
		Type:    vppinterfaces.Interface_MEMIF,
		Enabled: true,
		Link: &vppinterfaces.Interface_Memif{
			Memif: &vppinterfaces.MemifLink{
				Master: false,
			},
		},
	})
	return next.Server(ctx).Close(ctx, conn)
}
