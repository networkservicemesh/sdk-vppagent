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

package vxlan_test

import (
	"context"
	"io/ioutil"
	"net"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	vxlan_mechanism "github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/vxlan"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/checkvppagentmechanism"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/vxlan"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

func TestVxlanClient(t *testing.T) {
	// Turn off log output
	logrus.SetOutput(ioutil.Discard)
	srcIP := net.ParseIP("1.1.1.1")
	dstIP := net.ParseIP("1.1.1.2")
	testRequest := &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Id: "ConnectionId",
			Mechanism: &networkservice.Mechanism{
				Cls:  cls.REMOTE,
				Type: vxlan_mechanism.MECHANISM,
				Parameters: map[string]string{
					vxlan_mechanism.SrcIP: srcIP.String(),
					vxlan_mechanism.DstIP: dstIP.String(),
					vxlan_mechanism.VNI:   "2",
				},
			},
		},
	}
	vmech := vxlan_mechanism.ToMechanism(testRequest.GetConnection().GetMechanism())
	testConnToClose := testRequest.GetConnection()
	suite.Run(t, checkvppagentmechanism.NewClientSuite(
		vxlan.NewClient(srcIP, vxlan.EmptyInitFunc),
		vxlan_mechanism.MECHANISM,
		func(t *testing.T, mechanism *networkservice.Mechanism) {
			m := vxlan_mechanism.ToMechanism(mechanism)
			require.NotNil(t, m)
			assert.Equal(t, srcIP, m.SrcIP())
		},
		func(t *testing.T, conf *configurator.Config) { // Check the vppConfig
			// Basic Checks
			vppConfig := conf.GetVppConfig()
			vppInterfaces := vppConfig.GetInterfaces()
			require.Greater(t, len(vppInterfaces), 0)
			vppInterface := vppInterfaces[len(vppInterfaces)-1]
			assert.NotNil(t, vppInterface)

			// Check vxlan parameters
			vxlanInterface := vppInterface.GetVxlan()
			assert.NotNil(t, vxlanInterface)
			assert.Equal(t, srcIP.String(), vxlanInterface.GetSrcAddress())
			assert.Equal(t, dstIP.String(), vxlanInterface.GetDstAddress())
			vni := vmech.VNI()
			assert.Equal(t, vni, vxlanInterface.GetVni())
		},
		testRequest,
		testConnToClose,
	))
	t.Run("InvalidVNI", func(t *testing.T) {
		req := testRequest.Clone()
		req.GetConnection().GetMechanism().GetParameters()[vxlan_mechanism.VNI] = InvalidVNI
		clientUnderTest := vxlan.NewClient(srcIP, vxlan.EmptyInitFunc)
		conn, err := clientUnderTest.Request(vppagent.WithConfig(context.Background()), req)
		assert.Nil(t, conn)
		assert.NotNil(t, err)
		_, err = clientUnderTest.Close(vppagent.WithConfig(context.Background()), req.GetConnection())
		assert.NotNil(t, err)
	})
}
