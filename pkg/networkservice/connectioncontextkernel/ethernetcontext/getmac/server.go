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

// Package getmac provides networkservice chain elements for getting the mac address on kernel interfaces
package getmac

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/ligato/vpp-agent/api/models/linux"
	"github.com/networkservicemesh/api/pkg/api/connection"
	"github.com/networkservicemesh/api/pkg/api/connection/mechanisms/kernel"
	"github.com/networkservicemesh/api/pkg/api/connectioncontext"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

// NewServer creates a NetworkServiceServer chain element to set the EthernetContext for Kernel connection request
func NewServer(сс *grpc.ClientConn) networkservice.NetworkServiceServer {
	return &getMacKernelServer{
		client: configurator.NewConfiguratorClient(сс),
	}
}

type getMacKernelServer struct {
	client configurator.ConfiguratorClient
}

func (s *getMacKernelServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*connection.Connection, error) {
	var dstInterface *linux.Interface
	config := vppagent.Config(ctx)
	if mechanism := kernel.ToMechanism(request.GetConnection().GetMechanism()); mechanism != nil && len(config.GetLinuxConfig().GetInterfaces()) > 0 {
		dstInterface = config.LinuxConfig.Interfaces[len(config.LinuxConfig.Interfaces)-1]
	}
	conn, err := next.Server(ctx).Request(ctx, request)
	if err == nil && dstInterface != nil {
		dump, dumpErr := s.client.Dump(context.Background(), &configurator.DumpRequest{})
		if dumpErr != nil {
			logrus.Errorf("An error during ConfiguratorClient.Dump, err: %v", dumpErr.Error())
			return conn, err
		}
		for _, iface := range dump.Dump.LinuxConfig.Interfaces {
			if iface.Name == dstInterface.Name {
				request.GetConnection().GetContext().EthernetContext = &connectioncontext.EthernetContext{
					DstMac: iface.PhysAddress,
				}
				break
			}
		}
	}
	return conn, err
}

func (s *getMacKernelServer) Close(ctx context.Context, conn *connection.Connection) (*empty.Empty, error) {
	return next.Server(ctx).Close(ctx, conn)
}