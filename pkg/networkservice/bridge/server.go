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

// Package bridge provides networkservice chain elements for plugging vWires into bridges
package bridge

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ligato/vpp-agent/api/configurator"
	l2 "github.com/ligato/vpp-agent/api/models/vpp/l2"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	
	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type bridgeServer struct {
	name string
}

// NewServer creates a NetworkServiceServer that will plug an incoming vWire into a bridge named 'name'
func NewServer(name string) networkservice.NetworkServiceServer {
	return &bridgeServer{name: name}
}

func (b *bridgeServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	conf := vppagent.Config(ctx)
	b.insertInterfaceIntoBridge(conf)
	return next.Server(ctx).Request(ctx, request)
}

func (b *bridgeServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	conf := vppagent.Config(ctx)
	b.insertInterfaceIntoBridge(conf)
	return next.Server(ctx).Close(ctx, conn)
}

func (b *bridgeServer) insertInterfaceIntoBridge(conf *configurator.Config) {
	if len(conf.GetVppConfig().GetInterfaces()) > 0 {
		conf.GetVppConfig().BridgeDomains = append(conf.GetVppConfig().BridgeDomains, &l2.BridgeDomain{
			Name:                b.name,
			Flood:               false,
			UnknownUnicastFlood: false,
			Forward:             true,
			Learn:               true,
			ArpTermination:      false,
			Interfaces: []*l2.BridgeDomain_Interface{
				{
					Name:                    conf.GetVppConfig().GetInterfaces()[len(conf.GetVppConfig().GetInterfaces())-1].GetName(),
					BridgedVirtualInterface: false,
				},
			},
		})
	}
}
