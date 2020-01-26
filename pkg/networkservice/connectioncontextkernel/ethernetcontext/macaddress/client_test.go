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

package macaddress

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ligato/vpp-agent/api/models/linux"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection/mechanisms/kernel"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connectioncontext"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

func TestClientBasic(t *testing.T) {
	request := &networkservice.NetworkServiceRequest{
		Connection: &connection.Connection{
			Id: "1",
			Mechanism: &connection.Mechanism{
				Type: kernel.MECHANISM,
			},
			Context: &connectioncontext.ConnectionContext{
				EthernetContext: &connectioncontext.EthernetContext{
					DstMac: "0a-1b-3c-4d-5e-6f",
				},
			},
		},
	}
	server := next.NewNetworkServiceClient(vppagent.NewClient(), &testingClient{t}, NewClient())
	_, _ = server.Request(context.Background(), request)
	_, _ = server.Close(context.Background(), request.Connection)
}

type testingClient struct {
	*testing.T
}

func (t *testingClient) Request(ctx context.Context, in *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*connection.Connection, error) {
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
	conn, err := next.Client(ctx).Request(ctx, in, opts...)
	assert.Nil(t, err)
	assert.Equal(t, targetInterface.PhysAddress, conn.GetContext().GetEthernetContext().GetSrcMac())
	return conn, err
}

func (t *testingClient) Close(ctx context.Context, conn *connection.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
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
	result, err := next.Client(ctx).Close(ctx, conn, opts...)
	assert.Nil(t, err)
	assert.Equal(t, targetInterface.PhysAddress, conn.GetContext().GetEthernetContext().GetSrcMac())
	return result, err
}
