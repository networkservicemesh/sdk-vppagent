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

// Package checkkernelmechanism - provides a test suite for testing vppagent kernel mechanism implementation chain elements
package checkkernelmechanism

import (
	"context"
	"testing"

	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms/checkmechanism"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/adapters"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"
	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/checks/checkerror"
	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/null"

	vppagent_mechanism_suites "github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/checkvppagentmechanism"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

// ClientSuite - checkmechanism.ClientSuite extended to include additional Test Cases for vppagent kernel mechanisms
type ClientSuite struct {
	*checkmechanism.ClientSuite
	clientInodeNotFoundError networkservice.NetworkServiceClient
}

// NewClientSuite - return a test suite for testing vppagent mechanism implementations
//                  clientFactory - a function that takes a function to convert inodeNum to filenames and returns a client to be tested
//                  vppOnReturnCheck - function to check that the vppagent *configurator.Config *after* clientUnderTest has returned are correct
//                  request - NetworkServiceRequest to be used for testing
//                  connClose - Connection to be used for testing Close(...)
func NewClientSuite(
	clientFactory func(fileNameFromInodeNumberFunc func(string) (string, error)) networkservice.NetworkServiceClient,
	vppOnReturnCheck func(*testing.T, *configurator.Config),
	request *networkservice.NetworkServiceRequest,
	connClose *networkservice.Connection,
) *ClientSuite {
	return &ClientSuite{
		ClientSuite: vppagent_mechanism_suites.NewClientSuite(
			clientFactory(fileNameFromInodeNumberFunc),
			kernel.MECHANISM,
			func(t *testing.T, mechanism *networkservice.Mechanism) {},
			vppOnReturnCheck,
			request,
			connClose,
		),
		clientInodeNotFoundError: clientFactory(fileNameFromInodeNumberFuncError),
	}
}

// TestInodeNotFoundError - test case to test that an InodeNotFoundError results in the clientUnderTest returning an error
func (m *ClientSuite) TestInodeNotFoundError() {
	check := CheckClientInodeNotFoundError(m.T(), m.clientInodeNotFoundError)
	_, err := check.Request(vppagent.WithConfig(context.Background()), m.Request.Clone())
	assert.NotNil(m.T(), err)
	_, err = check.Close(vppagent.WithConfig(context.Background()), m.ConnClose.Clone())
	assert.NotNil(m.T(), err)
}

func fileNameFromInodeNumberFunc(inodeNum string) (string, error) {
	return "/proc/12/ns/net", nil
}

func fileNameFromInodeNumberFuncError(inodeNum string) (string, error) {
	return "", errors.New("Error finding inode")
}

// CheckClientInodeNotFoundError - check to make sure that if we get an error for not finding an Inode that it results
//                                 in an error being returned
//                                 t - *testing.T for the checks
//                                 clientUnderTest - the client being tested
func CheckClientInodeNotFoundError(t *testing.T, clientUnderTest networkservice.NetworkServiceClient) networkservice.NetworkServiceClient {
	return chain.NewNetworkServiceClient(
		checkerror.NewClient(t, false),
		clientUnderTest,
		adapters.NewServerToClient(
			mechanisms.NewServer(
				map[string]networkservice.NetworkServiceServer{
					kernel.MECHANISM: null.NewServer(),
				},
			),
		),
	)
}
