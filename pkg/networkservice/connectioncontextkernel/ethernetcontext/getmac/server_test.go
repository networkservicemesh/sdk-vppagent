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

package getmac

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/ligato/vpp-agent/api/models/linux"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestServerBasic(t *testing.T) {
	request := &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Id: "1",
			Mechanism: &networkservice.Mechanism{
				Type: kernel.MECHANISM,
			},
			Context: &networkservice.ConnectionContext{
				IpContext: &networkservice.IPContext{
					DstIpAddr: "172.16.1.2",
				},
			},
		},
	}

	server := next.NewNetworkServiceServer(
		vppagent.NewServer(),
		&testingServer{t},
		&getMacKernelServer{
			client: &testDumpConfiguratorClient{},
		})
	cc, err := server.Request(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, cc)
}

type testDumpConfiguratorClient struct {
}

func (t *testDumpConfiguratorClient) Get(ctx context.Context, in *configurator.GetRequest, opts ...grpc.CallOption) (*configurator.GetResponse, error) {
	panic("implement me")
}

func (t *testDumpConfiguratorClient) Update(ctx context.Context, in *configurator.UpdateRequest, opts ...grpc.CallOption) (*configurator.UpdateResponse, error) {
	panic("implement me")
}

func (t *testDumpConfiguratorClient) Delete(ctx context.Context, in *configurator.DeleteRequest, opts ...grpc.CallOption) (*configurator.DeleteResponse, error) {
	panic("implement me")
}

func (t *testDumpConfiguratorClient) Dump(ctx context.Context, in *configurator.DumpRequest, opts ...grpc.CallOption) (*configurator.DumpResponse, error) {
	return &configurator.DumpResponse{
		Dump: &configurator.Config{
			LinuxConfig: &linux.ConfigData{
				Interfaces: []*linux.Interface{
					{
						Name:        "DST-1-veth",
						PhysAddress: "1a-1b-3c-4d-5e-6f",
					},
					{
						Name:        "DST-1",
						PhysAddress: "0a-1b-3c-4d-5e-6f",
					},
					{
						Name:        "SRC-1-veth",
						PhysAddress: "2a-1b-3c-4d-5e-6f",
					},
					{
						Name:        "SRC-1",
						PhysAddress: "3a-1b-3c-4d-5e-6f",
					},
				},
			},
		},
	}, nil
}

func (t *testDumpConfiguratorClient) Notify(ctx context.Context, in *configurator.NotificationRequest, opts ...grpc.CallOption) (configurator.Configurator_NotifyClient, error) {
	panic("implement me")
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
				Name: "DST-1-veth",
			},
			{
				Name: "DST-1",
			},
		},
	}
	conn, err := next.Server(ctx).Request(ctx, in)
	assert.Nil(t, err)
	assert.NotNil(t, conn.GetContext().GetEthernetContext())
	assert.Equal(t, conn.GetContext().GetEthernetContext().GetDstMac(), "0a-1b-3c-4d-5e-6f")
	return conn, err
}

func (t *testingServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	return new(empty.Empty), nil
}
