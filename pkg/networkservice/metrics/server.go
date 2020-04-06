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

// Package metrics - implement vpp based metrics collector service, it update connection on passing Request() with set of new metrics received during interval
package metrics

import (
	"context"
	"errors"
	"fmt"

	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/serialize"
	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type metricsServer struct {
	executor  serialize.Executor
	vppClient configurator.StatsPollerServiceClient
}

func (s *metricsServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	conf := vppagent.Config(ctx)
	if conf == nil {
		return nil, errors.New("VPPAgent config is missing")
	}
	ifaces := conf.GetVppConfig().GetInterfaces()
	if len(ifaces) == 0 {
		return nil, errors.New("VPPAgent config should contain at least one interface")
	}
	index := request.GetConnection().GetPath().GetIndex()
	conn, err := next.Server(ctx).Request(ctx, request)
	if err == nil {
		<-s.executor.AsyncExec(func() {
			s.retieveVppStats(ctx, conn, index, ifaces)
		})
	}
	return conn, err
}

func (s *metricsServer) retieveVppStats(ctx context.Context, conn *networkservice.Connection, index uint32, ifaces []*vpp_interfaces.Interface) {
	logrus.Debugf("MetricsServer: Request Metrics")
	req := &configurator.PollStatsRequest{
		PeriodSec: 0,
		NumPolls:  0,
	}
	streamCtx, cancelOp := context.WithCancel(ctx)
	defer cancelOp()
	stream, err := s.vppClient.PollStats(streamCtx, req)
	if err != nil {
		logrus.Errorf("MetricsServer: Request Metrics: PollStats err: %v", err)
		return
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			logrus.Errorf("MetricsServer: stream.Recv() err: %v", err)
		} else {
			vppStats := resp.GetStats().GetVppStats()
			if vppStats.Interface != nil && vppStats.Interface.Name == ifaces[0].Name {
				conn.GetPath().GetPathSegments()[index].Metrics = s.newStatistics(vppStats.Interface)
				return
			}
			logrus.Debugf("MetricsServer: GetStats(): %v", vppStats)
		}
	}
}

func (s *metricsServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	return next.Server(ctx).Close(ctx, conn)
}

// NewServer creates a new metrics collector instance
func NewServer(vppClient configurator.StatsPollerServiceClient) networkservice.NetworkServiceServer {
	rv := &metricsServer{
		vppClient: vppClient,
	}
	return rv
}

func (s *metricsServer) newStatistics(stats *vpp_interfaces.InterfaceStats) map[string]string {
	metrics := make(map[string]string)
	metrics["rx_bytes"] = fmt.Sprint(stats.Rx.Bytes)
	metrics["tx_bytes"] = fmt.Sprint(stats.Tx.Bytes)
	metrics["rx_packets"] = fmt.Sprint(stats.Rx.Packets)
	metrics["tx_packets"] = fmt.Sprint(stats.Tx.Packets)
	metrics["rx_error_packets"] = fmt.Sprint(stats.RxError)
	metrics["tx_error_packets"] = fmt.Sprint(stats.TxError)
	return metrics
}
