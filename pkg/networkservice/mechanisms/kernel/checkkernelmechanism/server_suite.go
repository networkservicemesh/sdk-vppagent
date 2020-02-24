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

package checkkernelmechanism

import (
	"context"
	"testing"

	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/stretchr/testify/assert"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms/checkmechanism"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"
	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/checks/checkerror"

	vppagent_mechanism_suites "github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/checkvppagentmechanism"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

// ServerSuite - checkmechanism.ServerSuite extended to include additional Test Cases for vppagent kernel mechanisms
type ServerSuite struct {
	*checkmechanism.ServerSuite
	serverInodeNotFoundError networkservice.NetworkServiceServer
}

// NewServerSuite - returns a ServerSuite to test a vppagent kernel mechanism implementation
//                  serverFactory - a function that takes a function to convert inodeNum to filenames and returns a server to be tested
//                  vppOnReturnCheck - function to check that the vppagent *configurator.Config *after* clientUnderTest has returned are correct
//                  request - NetworkServiceRequest to be used for testing
//                  connClose - Connection to be used for testing Close(...)
func NewServerSuite(
	serverFactory func(fileNameFromInodeNumberFunc func(string) (string, error)) networkservice.NetworkServiceServer,
	vppCheck func(*testing.T, *configurator.Config),
	request *networkservice.NetworkServiceRequest,
	connClose *networkservice.Connection,
) *ServerSuite {
	return &ServerSuite{
		ServerSuite: vppagent_mechanism_suites.NewServerSuite(
			serverFactory(fileNameFromInodeNumberFunc),
			kernel.MECHANISM,
			func(t *testing.T, mechanism *networkservice.Mechanism) {},
			vppCheck,
			request,
			connClose,
		),
		serverInodeNotFoundError: serverFactory(fileNameFromInodeNumberFuncError),
	}
}

// TestInodeNotFoundError - test case to test that an InodeNotFoundError results in the clientUnderTest returning an error
func (m *ServerSuite) TestInodeNotFoundError() {
	check := CheckServerInodeNotFoundError(m.T(), m.serverInodeNotFoundError)
	_, err := check.Request(vppagent.WithConfig(context.Background()), m.Request.Clone())
	assert.NotNil(m.T(), err)
	_, err = check.Close(vppagent.WithConfig(context.Background()), m.ConnClose.Clone())
	assert.NotNil(m.T(), err)
}

// CheckServerInodeNotFoundError - check to make sure that if we get an error for not finding an Inode that it results
//                                 in an error being returned
//                                 t - *testing.T for the checks
//                                 serverUnderTest - the client being tested
func CheckServerInodeNotFoundError(t *testing.T, serverUnderTest networkservice.NetworkServiceServer) networkservice.NetworkServiceServer {
	return chain.NewNetworkServiceServer(
		checkerror.NewServer(t, false),
		serverUnderTest,
	)
}
