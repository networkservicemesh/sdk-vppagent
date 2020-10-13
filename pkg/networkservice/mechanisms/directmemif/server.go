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
	"net/url"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/memif"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type directMemifServer struct{}

// NewServer creates new direct memif server
func NewServer() networkservice.NetworkServiceServer {
	return NewServerWithNetwork("unixpacket")
}

// NewServerWithNetwork creates new direct memif server with specific network
func NewServerWithNetwork(net string) networkservice.NetworkServiceServer {
	return &directMemifServer{}
}

func (d *directMemifServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	if mechanism := memif.ToMechanism(request.GetConnection().GetMechanism()); mechanism != nil {
		vc := vppagent.Config(ctx).GetVppConfig()
		l := len(vc.GetInterfaces())
		if l < 2 {
			return next.Server(ctx).Request(ctx, request)
		}
		client := vc.GetInterfaces()[l-2]
		endpoint := vc.GetInterfaces()[l-1]
		if client.GetMemif() == nil || endpoint.GetMemif() == nil {
			return next.Server(ctx).Request(ctx, request)
		}
		vc.Interfaces = vc.GetInterfaces()[:l-2]
		mechanism.SetSocketFileURL((&url.URL{Scheme: "file", Path: endpoint.GetMemif().GetSocketFilename()}).String())
	}

	return next.Server(ctx).Request(ctx, request)
}

func (d *directMemifServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	if mechanism := memif.ToMechanism(conn.GetMechanism()); mechanism != nil {
		vc := vppagent.Config(ctx).GetVppConfig()
		l := len(vc.GetInterfaces())
		if l < 2 {
			return next.Server(ctx).Close(ctx, conn)
		}
		client := vc.GetInterfaces()[l-2]
		endpoint := vc.GetInterfaces()[l-1]
		if client.GetMemif() == nil || endpoint.GetMemif() == nil {
			return next.Server(ctx).Close(ctx, conn)
		}
		vc.Interfaces = vc.GetInterfaces()[:l-2]
		mechanism.SetSocketFileURL((&url.URL{Scheme: "file", Path: endpoint.GetMemif().GetSocketFilename()}).String())
	}
	return next.Server(ctx).Close(ctx, conn)
}
