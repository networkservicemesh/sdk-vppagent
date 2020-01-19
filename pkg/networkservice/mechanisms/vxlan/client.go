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

// Package vxlan provides networkservice chain elements that support the vxlan Mechanism
package vxlan

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ligato/vpp-agent/api/models/vpp"
	vppinterfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection/mechanisms/vxlan"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type vxlanClient struct{}

// NewClient provides a NetworkServiceClient chain elements that support the vxlan Mechanism
func NewClient() networkservice.NetworkServiceClient {
	return &vxlanClient{}
}

func (v *vxlanClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*connection.Connection, error) {
	if err := v.appendInterfaceConfig(ctx, request.GetConnection()); err != nil {
		return nil, err
	}
	return next.Client(ctx).Request(ctx, request, opts...)
}

func (v *vxlanClient) Close(ctx context.Context, conn *connection.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	if err := v.appendInterfaceConfig(ctx, conn); err != nil {
		return nil, err
	}
	return next.Client(ctx).Close(ctx, conn, opts...)
}

func (v *vxlanClient) appendInterfaceConfig(ctx context.Context, conn *connection.Connection) error {
	conf := vppagent.Config(ctx)
	if mechanism := vxlan.ToMechanism(conn.GetMechanism()); mechanism != nil {
		srcIP, err := mechanism.SrcIP()
		if err != nil {
			return err
		}
		dstIP, err := mechanism.DstIP()
		if err != nil {
			return err
		}
		vni, err := mechanism.VNI()
		if err != nil {
			return err
		}
		conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().Interfaces, &vpp.Interface{
			Name:    conn.GetId(),
			Type:    vppinterfaces.Interface_VXLAN_TUNNEL,
			Enabled: true,
			Link: &vppinterfaces.Interface_Vxlan{
				Vxlan: &vppinterfaces.VxlanLink{
					SrcAddress: dstIP,
					DstAddress: srcIP,
					Vni:        vni,
				},
			},
		})
	}
	return nil
}
