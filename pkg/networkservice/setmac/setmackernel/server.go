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

// Package setmackernel provides networkservice chain elements for setting the mac address on kernel interfaces
package setmackernel

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection/mechanisms/kernel"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type setVppMacServer struct{}

// NewServer creates a NetworkServiceServer chain element to set the mac address on a kernel interface
// It sets the Mac Address on the *kernel* side of an interface plugged into the
// Endpoint.  Generally only used by privileged Endpoints like those implementing
// the Cross Connect Network Service for K8s (formerly known as NSM Forwarder).
//                                         Endpoint
//                              +---------------------------+
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//          +-------------------+                           |
// setmackerel.NewServer()      |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              +---------------------------+
//
func NewServer() networkservice.NetworkServiceServer {
	return &setVppMacServer{}
}

func (s *setVppMacServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*connection.Connection, error) {
	conf := vppagent.Config(ctx)
	conn, err := next.Client(ctx).Request(ctx, request)
	if err != nil {
		return nil, err
	}
	if mechanism := kernel.ToMechanism(request.GetConnection().GetMechanism()); mechanism != nil {
		index := len(conf.GetLinuxConfig().GetInterfaces()) - 1
		conf.GetLinuxConfig().GetInterfaces()[index+1].PhysAddress = conn.GetContext().GetEthernetContext().GetDstMac()
	}
	return conn, nil
}

func (s *setVppMacServer) Close(ctx context.Context, conn *connection.Connection) (*empty.Empty, error) {
	conf := vppagent.Config(ctx)
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil && len(conf.GetLinuxConfig().GetInterfaces()) > 0 {
		conf.GetLinuxConfig().GetInterfaces()[len(conf.GetVppConfig().GetInterfaces())-1].PhysAddress = conn.GetContext().GetEthernetContext().GetDstMac()
	}
	return next.Client(ctx).Close(ctx, conn)
}
