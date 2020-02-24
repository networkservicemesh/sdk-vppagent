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

package checkvppagentmechanism

import (
	"context"
	"testing"

	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms/checkmechanism"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

// NewServerSuite - returns a ServerSuite to test a vppagent mechanism implementation
//                  serverUnderTest - the server being tested to make sure it correctly implements a Mechanism
//                  configureContext - a function that is applied to context.Background to make sure anything
//                                     needed by the serverUnderTest is present in the context.Context
//                  mechanismType - Mechanism.Type implemented by the serverUnderTest
//                  vppOnReturnCheck - function to check that the vppagent *configurator.Config *after* clientUnderTest has returned are correct
//                  request - NetworkServiceRequest to be used for testing
//                  connClose - Connection to be used for testing Close(...)
func NewServerSuite(
	serverUnderTest networkservice.NetworkServiceServer,
	vppCheckOnReturn func(*testing.T, *configurator.Config),
	request *networkservice.NetworkServiceRequest,
	connClose *networkservice.Connection,
) *checkmechanism.ServerSuite {
	return checkmechanism.NewServerSuite(
		serverUnderTest,
		vppagent.WithConfig,
		kernel.MECHANISM,
		func(t *testing.T, ctx context.Context) {
			conf := vppagent.Config(ctx)
			require.NotNil(t, conf)
			require.NotNil(t, conf.GetVppConfig())
			require.NotNil(t, conf.GetLinuxConfig())
			require.NotNil(t, conf.GetNetallocConfig())
			vppCheckOnReturn(t, conf)
		},
		request,
		connClose,
	)
}
