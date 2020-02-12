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

package vxlan

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ligato/vpp-agent/api/models/vpp"
	vppinterfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/vxlan"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type vxlanServer struct{}

// NewServer provides a NetworkServiceServer chain elements that support the vxlan Mechanism
func NewServer() networkservice.NetworkServiceServer {
	return &vxlanServer{}
}

func (v *vxlanServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	if err := v.appendInterfaceConfig(ctx, request.GetConnection()); err != nil {
		return nil, err
	}
	return next.Server(ctx).Request(ctx, request)
}

func (v *vxlanServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	if err := v.appendInterfaceConfig(ctx, conn); err != nil {
		return nil, err
	}
	return next.Server(ctx).Close(ctx, conn)
}

func (v *vxlanServer) appendInterfaceConfig(ctx context.Context, conn *networkservice.Connection) error {
	conf := vppagent.Config(ctx)
	if mechanism := vxlan.ToMechanism(conn.GetMechanism()); mechanism != nil {
		// TODO do VNI selection here
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
					// Note: srcIP and Dst Ip are relative to the *client*, and so on the server side are flipped
					SrcAddress: mechanism.DstIP().String(),
					DstAddress: mechanism.SrcIP().String(),
					Vni:        vni,
				},
			},
		})
	}
	return nil
}
