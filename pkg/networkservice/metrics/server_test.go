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

package metrics_test

import (
	"context"
	"testing"
	"time"

	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vppInt "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/metrics"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type testClient struct {
	notifications chan *configurator.PollStatsResponse
	stream        *testClientStream
}

func (t *testClient) PollStats(ctx context.Context, in *configurator.PollStatsRequest, opts ...grpc.CallOption) (configurator.StatsPollerService_PollStatsClient, error) {
	t.stream = &testClientStream{
		request: in,
		client:  t,
		ctx:     ctx,
	}
	return t.stream, nil
}

type testClientStream struct {
	client *testClient
	configurator.StatsPollerService_PollStatsClient
	request *configurator.PollStatsRequest
	ctx     context.Context
}

func (s *testClientStream) Recv() (*configurator.PollStatsResponse, error) {
	n := <-s.client.notifications
	return n, nil
}

func TestMonitorVppEvents(t *testing.T) {
	client := &testClient{
		notifications: make(chan *configurator.PollStatsResponse, 10),
	}
	server := metrics.NewServer(client)
	require.NotNil(t, server)

	ctx := vppagent.WithConfig(context.Background())
	config := vppagent.Config(ctx)

	config.VppConfig.Interfaces = append(config.VppConfig.Interfaces, &vppInt.Interface{
		Name: "client-id0",
	})

	client.notifications <- createDummyNotification()
	req := newRequest()
	response, err := server.Request(ctx, req)
	require.NotNil(t, response)
	require.Nil(t, err)

	// Check metrics returned.
	require.Equal(t, "11", response.GetPath().GetPathSegments()[0].GetMetrics()["rx_bytes"])
	require.Equal(t, 6, len(response.GetPath().GetPathSegments()[0].GetMetrics()))

	// Check if context has do deadline.
	_, ok := client.stream.ctx.Deadline()
	require.Equal(t, false, ok)

	_, err = server.Close(ctx, response)
	require.Nil(t, err)

	select {
	// Check if cancel is passed to context.
	case <-client.stream.ctx.Done():
	case <-time.After(1 * time.Second):
	}
	// Check if collect go routing is terminated
	require.NotNil(t, client.stream.ctx.Err())
}

func newRequest() *networkservice.NetworkServiceRequest {
	return &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Id: "id0",
			Path: &networkservice.Path{
				PathSegments: []*networkservice.PathSegment{
					{
						Name: "qwe",
					},
				},
			},
		},
	}
}

func createDummyNotification() *configurator.PollStatsResponse {
	vppStats := &configurator.Stats_VppStats{
		VppStats: &vpp.Stats{
			Interface: &vppInt.InterfaceStats{
				Name:    "client-id0",
				Rx:      &vppInt.InterfaceStats_CombinedCounter{Bytes: 11, Packets: 11},
				Tx:      &vppInt.InterfaceStats_CombinedCounter{Bytes: 12, Packets: 12},
				RxError: uint64(time.Now().Second()),
				TxError: 0,
				RxNoBuf: 0,
				RxMiss:  0,
				Drops:   12,
				Punts:   0,
				Ip4:     13,
				Ip6:     14,
				Mpls:    100,
			},
		},
	}
	return &configurator.PollStatsResponse{
		PollSeq: 0,
		Stats: &configurator.Stats{
			Stats: vppStats,
		},
	}
}
