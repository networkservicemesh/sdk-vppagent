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

// Package directmemif provides server chain element that create connection between two memif interfaces
package directmemif

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/serialize"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/directmemif/proxy"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type directMemifServer struct {
	net      string
	executor serialize.Executor
	proxies  map[string]proxy.Proxy
}

// NewServer creates new direct memif server
func NewServer() networkservice.NetworkServiceServer {
	return NewServerWithNetwork("unixpacket")
}

// NewServerWithNetwork creates new direct memif server with specific network
func NewServerWithNetwork(net string) networkservice.NetworkServiceServer {
	return &directMemifServer{
		executor: serialize.NewExecutor(),
		proxies:  map[string]proxy.Proxy{},
		net:      net,
	}
}

func (d *directMemifServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	conn, err := next.Server(ctx).Request(ctx, request)
	if err != nil {
		return conn, err
	}
	c := vppagent.Config(ctx)
	l := len(c.GetVppConfig().GetInterfaces())
	if l < 2 {
		return conn, err
	}
	client := c.GetVppConfig().GetInterfaces()[l-2]
	endpoint := c.GetVppConfig().GetInterfaces()[l-1]
	if client.GetMemif() == nil || endpoint.GetMemif() == nil {
		return conn, err
	}
	c.GetVppConfig().Interfaces = c.GetVppConfig().GetInterfaces()[:l-2]
	if err != nil {
		return nil, err
	}
	p, err := proxy.New(client.GetMemif().GetSocketFilename(), endpoint.GetMemif().GetSocketFilename(), d.net, proxy.StopListenerAdapter(func() {
		d.executor.AsyncExec(func() {
			delete(d.proxies, request.Connection.Id)
		})
	}))
	if err != nil {
		return nil, err
	}
	d.executor.AsyncExec(func() {
		prev := d.proxies[conn.Id]
		if prev != nil {
			_ = prev.Stop()
			d.executor.AsyncExec(func() {
				d.proxies[conn.Id] = p
			})
		} else {
			d.proxies[conn.Id] = p
		}
	})
	err = p.Start()
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (d *directMemifServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	c := vppagent.Config(ctx)
	r, err := next.Server(ctx).Close(ctx, conn)
	if err != nil {
		return r, err
	}
	l := len(c.GetVppConfig().GetInterfaces())
	if l < 2 {
		return r, err
	}
	client := c.GetVppConfig().GetInterfaces()[l-2]
	endpoint := c.GetVppConfig().GetInterfaces()[l-1]
	if client.GetMemif() == nil || endpoint.GetMemif() == nil {
		return r, err
	}
	d.executor.AsyncExec(func() {
		p := d.proxies[conn.Id]
		if p != nil {
			_ = d.proxies[conn.Id].Stop()
			delete(d.proxies, conn.Id)
		}
	})
	return r, err
}
