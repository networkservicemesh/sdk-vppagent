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
	"context"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/srv6"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/utils/checks/testinterfaceappender"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms/checkmechanism"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	srv6_mechanism "github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/srv6"
)

func TestSrv6Client(t *testing.T) {
	// Turn off log output
	logrus.SetOutput(ioutil.Discard)
	parameters := configureTestSRv6Parameters()
	localInterfaceName := "client-ConnectionId"
	testMechanism := configureTestSRv6Mechanism(parameters)
	testRequest := &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Id:        "ConnectionId",
			Mechanism: testMechanism,
		},
	}
	c := next.NewNetworkServiceClient(
		testinterfaceappender.NewClient(),
		srv6.NewClient(),
	)
	testConnToClose := testRequest.GetConnection()
	suite.Run(t, checkmechanism.NewClientSuite(
		c,
		vppagent.WithConfig,
		srv6_mechanism.MECHANISM,
		func(t *testing.T, mechanism *networkservice.Mechanism) {
			m := srv6_mechanism.ToMechanism(mechanism)
			require.NotNil(t, m)
		},
		func(t *testing.T, ctx context.Context) {
			conf := vppagent.Config(ctx)
			require.NotNil(t, conf)
			require.NotNil(t, conf.GetVppConfig())
			require.NotNil(t, conf.GetLinuxConfig())
			require.NotNil(t, conf.GetNetallocConfig())

			// Basic Checks
			vppConfig := conf.GetVppConfig()
			vppInterfaces := vppConfig.GetInterfaces()
			require.Len(t, vppInterfaces, 1)
			vppInterface := vppInterfaces[len(vppInterfaces)-1]
			assert.NotNil(t, vppInterface)

			// Checks that sets the correct configs
			assert.Equal(t, expectedVppConfigSrv6Localsids(parameters, localInterfaceName), vppConfig.Srv6Localsids)
			assert.Equal(t, expectedVppConfigSrv6Policies(parameters), vppConfig.Srv6Policies)
			assert.Equal(t, expectedVppConfigSrv6Steerings(testRequest, parameters, localInterfaceName), vppConfig.Srv6Steerings)

			// Make sure who invokes the appendInterfaceConfig function:
			//   Request(...) with connect=true
			// or
			//   Close(...) with connect=false
			if len(vppConfig.Vrfs) == 1 {
				assert.Equal(t, expectedVppConfigVrfs(), vppConfig.Vrfs)
				assert.Equal(t, expectedVppConfigRoutes(parameters), vppConfig.Routes)
				assert.Equal(t, expectedVppConfigArps(parameters), vppConfig.Arps)
			}
		},
		func(t *testing.T, ctx context.Context) {},
		testRequest,
		testConnToClose,
	))
}
