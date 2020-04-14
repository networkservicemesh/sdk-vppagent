// Copyright (c) 2020 Cisco and/or its affiliates.
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

package srv6

import (
	"context"
	"math"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/srv6"
	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	vpp_srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

func appendInterfaceConfig(ctx context.Context, conn *networkservice.Connection, connect bool) error {
	conf := vppagent.Config(ctx)
	mechanism := srv6.ToMechanism(conn.GetMechanism())
	if mechanism == nil {
		return nil
	}
	vppConfig := conf.GetVppConfig()

	dstHostLocalSID := mechanism.DstHostLocalSID()
	if dstHostLocalSID == "" {
		return errors.New("destination host local SID is empty")
	}
	hardwareAddress := mechanism.DstHardwareAddress()
	if hardwareAddress == "" {
		return errors.New("source hardware address is empty")
	}
	srcBSID := mechanism.SrcBSID()
	if srcBSID == "" {
		return errors.New("source BSID is empty")
	}
	srcLocalSID := mechanism.SrcLocalSID()
	if srcLocalSID == "" {
		return errors.New("source local SID is empty")
	}
	dstLocalSID := mechanism.DstLocalSID()
	if dstLocalSID == "" {
		return errors.New("destination local SID is empty")
	}

	var localIfaceName string
	if ifaces := vppConfig.GetInterfaces(); len(ifaces) == 1 {
		localIfaceName = ifaces[0].Name
	} else {
		return errors.Errorf("failed to choose local interface for srv6 mechanism: %v", ifaces)
	}

	vppConfig.Srv6Localsids = []*vpp_srv6.LocalSID{
		{
			Sid: srcLocalSID,
			EndFunction: &vpp_srv6.LocalSID_EndFunctionDx2{
				EndFunctionDx2: &vpp_srv6.LocalSID_EndDX2{
					VlanTag:           math.MaxUint32,
					OutgoingInterface: localIfaceName,
				},
			},
		},
	}
	vppConfig.Srv6Policies = []*vpp_srv6.Policy{
		{
			Bsid: srcBSID,
			SegmentLists: []*vpp_srv6.Policy_SegmentList{
				{
					Segments: []string{
						dstHostLocalSID,
						dstLocalSID,
					},
					Weight: 0,
				},
			},
			SrhEncapsulation: true,
		},
	}

	vppConfig.Srv6Steerings = []*vpp_srv6.Steering{
		{
			Name: conn.GetId(),
			PolicyRef: &vpp_srv6.Steering_PolicyBsid{
				PolicyBsid: srcBSID,
			},
			Traffic: &vpp_srv6.Steering_L2Traffic_{
				L2Traffic: &vpp_srv6.Steering_L2Traffic{
					InterfaceName: localIfaceName,
				},
			},
		},
	}

	if connect {
		vppConfig.Vrfs = []*vpp_l3.VrfTable{
			{
				Id:       math.MaxUint32,
				Protocol: vpp_l3.VrfTable_IPV6,
				Label:    "SRv6 steering of IP6 prefixes through BSIDs",
			},
		}

		vppConfig.Routes = append(vppConfig.Routes, &vpp.Route{
			Type:              vpp_l3.Route_INTER_VRF,
			OutgoingInterface: "mgmt",
			DstNetwork:        dstHostLocalSID + "/128",
			Weight:            1,
			NextHopAddr:       dstHostLocalSID,
		})

		vppConfig.Arps = append(vppConfig.Arps, &vpp.ARPEntry{
			Interface:   "mgmt",
			IpAddress:   dstHostLocalSID,
			PhysAddress: hardwareAddress,
			Static:      true,
		})
	}

	return nil
}
