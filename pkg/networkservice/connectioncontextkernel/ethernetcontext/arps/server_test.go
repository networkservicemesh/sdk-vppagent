// Copyright (c) 2020 Doc.ai and/or its affiliates.
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

package arps

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ligato/vpp-agent/api/models/linux"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/stretchr/testify/assert"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

func TestClientBasic(t *testing.T) {
	request := &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Id: "1",
			Mechanism: &networkservice.Mechanism{
				Type: kernel.MECHANISM,
			},
			Context: &networkservice.ConnectionContext{
				EthernetContext: &networkservice.EthernetContext{
					DstMac: "0a-1b-3c-4d-5e-6f",
				},
				IpContext: &networkservice.IPContext{
					DstIpAddr: "172.16.1.2",
				},
			},
		},
	}
	server := next.NewNetworkServiceServer(vppagent.NewServer(), &testingServer{t}, NewServer())
	_, _ = server.Request(context.Background(), request)
	_, _ = server.Close(context.Background(), request.GetConnection())
}

type testingServer struct {
	*testing.T
}

func (t *testingServer) Request(ctx context.Context, in *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	config := vppagent.Config(ctx)
	assert.NotNil(t, config)
	config.LinuxConfig = &linux.ConfigData{
		Interfaces: []*linux.Interface{
			{
				Name: "client-1",
			},
		},
	}
	conn, err := next.Server(ctx).Request(ctx, in)
	assert.Nil(t, err)
	expectedArp := &linux.ARPEntry{
		Interface: config.LinuxConfig.Interfaces[0].Name,
		IpAddress: in.GetConnection().GetContext().GetIpContext().GetDstIpAddr(),
		HwAddress: in.GetConnection().GetContext().GetEthernetContext().GetDstMac(),
	}
	assert.Equal(t, 1, len(config.LinuxConfig.ArpEntries))
	assert.Equal(t, expectedArp, config.LinuxConfig.ArpEntries[0])
	return conn, err
}

func (t *testingServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	config := vppagent.Config(ctx)
	assert.NotNil(t, config)
	targetInterface := &linux.Interface{
		Name: "SRC-1",
	}
	config.LinuxConfig = &linux.ConfigData{
		Interfaces: []*linux.Interface{
			targetInterface,
		},
	}
	result, err := next.Server(ctx).Close(ctx, conn)
	assert.Nil(t, err)
	expectedArp := &linux.ARPEntry{
		Interface: config.LinuxConfig.Interfaces[0].Name,
		IpAddress: conn.GetContext().GetIpContext().GetDstIpAddr(),
		HwAddress: conn.GetContext().GetEthernetContext().GetDstMac(),
	}
	assert.Equal(t, 1, len(config.LinuxConfig.ArpEntries))
	assert.Equal(t, expectedArp, config.LinuxConfig.ArpEntries[0])
	return result, err
}
