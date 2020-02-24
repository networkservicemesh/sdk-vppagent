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

// Package checkvppagentmechanism - provides a test suite for testing vppagent mechanism implementation chain elements
package checkvppagentmechanism

import (
	"context"
	"testing"

	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms/checkmechanism"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

// NewClientSuite - return a test suite for testing vppagent mechanism implementations
//                  clientUnderTest - the client we are testing to make sure it correctly implements a Mechanism
//                  mechanismType - Mechanism.Type implemented by the clientUnderTest
//                  mechanismCheck - function to check that the clientUnderTest has properly added Mechanism.Parameters
//                                   to the Mechanism it has appended to MechanismPreferences
//                  vppOnReturnCheck - function to check that the vppagent *configurator.Config *after* clientUnderTest has returned are correct
//                  request - NetworkServiceRequest to be used for testing
//                  connClose - Connection to be used for testing Close(...)
func NewClientSuite(
	clientUnderTest networkservice.NetworkServiceClient,
	mechanismType string,
	mechanismCheck func(*testing.T, *networkservice.Mechanism),
	vppOnReturnCheck func(*testing.T, *configurator.Config),
	request *networkservice.NetworkServiceRequest,
	connClose *networkservice.Connection,
) *checkmechanism.ClientSuite {
	return checkmechanism.NewClientSuite(
		clientUnderTest,
		vppagent.WithConfig,
		mechanismType,
		mechanismCheck,
		func(t *testing.T, ctx context.Context) {
			conf := vppagent.Config(ctx)
			require.NotNil(t, conf)
			require.NotNil(t, conf.GetVppConfig())
			require.NotNil(t, conf.GetLinuxConfig())
			require.NotNil(t, conf.GetNetallocConfig())
			vppOnReturnCheck(t, conf)
		},
		func(t *testing.T, ctx context.Context) {
			conf := vppagent.Config(ctx)
			expectedConf := vppagent.Config(vppagent.WithConfig(context.Background()))
			assert.Equal(t, expectedConf, conf)
		},
		request,
		connClose,
	)
}
