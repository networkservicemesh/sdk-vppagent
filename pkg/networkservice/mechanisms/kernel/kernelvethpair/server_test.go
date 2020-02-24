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

package kernelvethpair_test

import (
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/kernel/checkkernelmechanism"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/kernel/kernelvethpair"
)

func TestKernelTapServer(t *testing.T) {
	// Turn off log output
	logrus.SetOutput(ioutil.Discard)
	mechanism := &networkservice.Mechanism{
		Cls:  cls.LOCAL,
		Type: kernel.MECHANISM,
		Parameters: map[string]string{
			kernel.NetNSInodeKey: "12",
		},
	}
	testRequest := &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Id:        "ConnectionId",
			Mechanism: mechanism,
		},
	}
	testConnToClose := &networkservice.Connection{
		Id:        "ConnectionId",
		Mechanism: mechanism,
	}
	kmech := kernel.ToMechanism(mechanism)
	mechanism.GetParameters()[kernel.InterfaceNameKey] = kmech.GetInterfaceName(testConnToClose)
	suite.Run(t, checkkernelmechanism.NewServerSuite(
		kernelvethpair.NewTestableServer,
		checkVppAgentConfig("server", testRequest),
		testRequest,
		testConnToClose,
	))
}
