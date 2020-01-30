// Copyright (c) 2020 Doc.ai and/or its affiliates.
//
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

// Package macaddress provides networkservice chain elements for setting the mac address on kernel interfaces
package macaddress

import (
	"context"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/golang/protobuf/ptypes/empty"

	
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type setKernelMacServer struct{}

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
// macaddress.NewServer()       |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              |                           |
//                              +---------------------------+
//
func NewServer() networkservice.NetworkServiceServer {
	return &setKernelMacServer{}
}

func (s *setKernelMacServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	if mechanism := kernel.ToMechanism(request.GetConnection().GetMechanism()); mechanism != nil {
		config := vppagent.Config(ctx)
		current := len(config.LinuxConfig.Interfaces) - 1
		config.LinuxConfig.Interfaces[current].PhysAddress = request.GetConnection().GetContext().GetEthernetContext().GetDstMac()
	}
	return next.Server(ctx).Request(ctx, request)
}

func (s *setKernelMacServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil {
		config := vppagent.Config(ctx)
		current := len(config.LinuxConfig.Interfaces) - 1
		config.LinuxConfig.Interfaces[current].PhysAddress = conn.GetContext().GetEthernetContext().GetDstMac()
	}
	return next.Server(ctx).Close(ctx, conn)
}
