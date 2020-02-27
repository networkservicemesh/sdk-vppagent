// Copyright (c) 2020 Cisco Systems, Inm.
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
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vppinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/vxlan"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type vxlanClient struct {
	srcIP net.IP
}

// NewClient provides a NetworkServiceClient chain elements that support the vxlan Mechanism
func NewClient(srcIP net.IP) networkservice.NetworkServiceClient {
	return &vxlanClient{
		srcIP: srcIP,
	}
}

func (v *vxlanClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	preferredMechanism := &networkservice.Mechanism{
		Cls:  cls.REMOTE,
		Type: vxlan.MECHANISM,
		Parameters: map[string]string{
			vxlan.SrcIP: v.srcIP.String(),
		},
	}
	request.MechanismPreferences = append(request.MechanismPreferences, preferredMechanism)
	rv, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if configErr := v.appendInterfaceConfig(ctx, request.GetConnection()); configErr != nil {
		return nil, configErr
	}
	return rv, err
}

func (v *vxlanClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	rv, err := next.Client(ctx).Close(ctx, conn, opts...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if configErr := v.appendInterfaceConfig(ctx, conn); configErr != nil {
		return nil, configErr
	}
	return rv, err
}

func (v *vxlanClient) appendInterfaceConfig(ctx context.Context, conn *networkservice.Connection) error {
	conf := vppagent.Config(ctx)
	if mechanism := vxlan.ToMechanism(conn.GetMechanism()); mechanism != nil {
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
					SrcAddress: mechanism.SrcIP().String(),
					DstAddress: mechanism.DstIP().String(),
					Vni:        vni,
				},
			},
		})
	}
	return nil
}
