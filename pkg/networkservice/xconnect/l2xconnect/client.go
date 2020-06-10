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

// Package l2xconnect provides a NetworkServiceClient chain element for an l2 cross connect
package l2xconnect

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type l2XconnectClient struct{}

// NewClient - creates a NetworkServiceClient chain element for an l2 cross connect
func NewClient() networkservice.NetworkServiceClient {
	return &l2XconnectClient{}
}

func (l *l2XconnectClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	rv, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	conf := vppagent.Config(ctx)
	l.appendL2XConnect(conf)
	return rv, nil
}

func (l *l2XconnectClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	rv, err := next.Client(ctx).Close(ctx, conn, opts...)
	if err != nil {
		return nil, err
	}
	conf := vppagent.Config(ctx)
	l.appendL2XConnect(conf)
	return rv, nil
}

func (l *l2XconnectClient) appendL2XConnect(conf *configurator.Config) {
	if len(conf.GetVppConfig().GetInterfaces()) >= 2 {
		ifaces := conf.GetVppConfig().GetInterfaces()[len(conf.GetVppConfig().Interfaces)-2:]
		conf.GetVppConfig().XconnectPairs = append(conf.GetVppConfig().XconnectPairs,
			&l2.XConnectPair{
				ReceiveInterface:  ifaces[0].Name,
				TransmitInterface: ifaces[1].Name,
			},
			&l2.XConnectPair{
				ReceiveInterface:  ifaces[1].Name,
				TransmitInterface: ifaces[0].Name,
			})
	}
}
