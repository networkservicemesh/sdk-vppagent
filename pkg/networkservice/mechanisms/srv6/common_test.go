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

package srv6_test

import (
	"math"

	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	vpp_srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	srv6_mechanism "github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/srv6"
)

func configureTestSRv6Mechanism(parameters map[string]string) *networkservice.Mechanism {
	return &networkservice.Mechanism{
		Cls:        cls.REMOTE,
		Type:       srv6_mechanism.MECHANISM,
		Parameters: parameters,
	}
}

func configureTestSRv6Parameters() map[string]string {
	return map[string]string{
		srv6_mechanism.DstHardwareAddress: "hardwareAddress",
		srv6_mechanism.SrcLocalSID:        "1:1:1:1:1:1:1:1",
		srv6_mechanism.DstHostLocalSID:    "1:1:1:1:1:1:1:2",
		srv6_mechanism.SrcBSID:            "1:1:1:1:1:1:1:3",
		srv6_mechanism.DstLocalSID:        "1:1:1:1:1:1:1:4",
	}
}

func expectedVppConfigSrv6Localsids(parameters map[string]string, localInterfaceName string) []*vpp_srv6.LocalSID {
	return []*vpp_srv6.LocalSID{
		{
			Sid: parameters[srv6_mechanism.SrcLocalSID],
			EndFunction: &vpp_srv6.LocalSID_EndFunctionDx2{
				EndFunctionDx2: &vpp_srv6.LocalSID_EndDX2{
					VlanTag:           math.MaxUint32,
					OutgoingInterface: localInterfaceName,
				},
			},
		},
	}
}

func expectedVppConfigSrv6Policies(parameters map[string]string) []*vpp_srv6.Policy {
	return []*vpp_srv6.Policy{
		{
			Bsid: parameters[srv6_mechanism.SrcBSID],
			SegmentLists: []*vpp_srv6.Policy_SegmentList{
				{
					Segments: []string{
						parameters[srv6_mechanism.DstHostLocalSID],
						parameters[srv6_mechanism.DstLocalSID],
					},
					Weight: 0,
				},
			},
			SrhEncapsulation: true,
		},
	}
}

func expectedVppConfigSrv6Steerings(
	testRequest *networkservice.NetworkServiceRequest,
	parameters map[string]string,
	localInterfaceName string,
) []*vpp_srv6.Steering {
	return []*vpp_srv6.Steering{
		{
			Name: testRequest.Connection.GetId(),
			PolicyRef: &vpp_srv6.Steering_PolicyBsid{
				PolicyBsid: parameters[srv6_mechanism.SrcBSID],
			},
			Traffic: &vpp_srv6.Steering_L2Traffic_{
				L2Traffic: &vpp_srv6.Steering_L2Traffic{
					InterfaceName: localInterfaceName,
				},
			},
		},
	}
}

func expectedVppConfigVrfs() []*vpp_l3.VrfTable {
	return []*vpp_l3.VrfTable{
		{
			Id:       math.MaxUint32,
			Protocol: vpp_l3.VrfTable_IPV6,
			Label:    "SRv6 steering of IP6 prefixes through BSIDs",
		},
	}
}

func expectedVppConfigRoutes(parameters map[string]string) []*vpp.Route {
	return []*vpp.Route{
		{
			Type:              vpp_l3.Route_INTER_VRF,
			OutgoingInterface: "mgmt",
			DstNetwork:        parameters[srv6_mechanism.DstHostLocalSID] + "/128",
			Weight:            1,
			NextHopAddr:       parameters[srv6_mechanism.DstHostLocalSID],
		},
	}
}

func expectedVppConfigArps(parameters map[string]string) []*vpp.ARPEntry {
	return []*vpp.ARPEntry{
		{
			Interface:   "mgmt",
			IpAddress:   parameters[srv6_mechanism.DstHostLocalSID],
			PhysAddress: parameters[srv6_mechanism.DstHardwareAddress],
			Static:      true,
		},
	}
}
