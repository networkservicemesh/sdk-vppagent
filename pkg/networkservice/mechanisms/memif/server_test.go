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

package memif_test

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"

	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/memif"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/checkvppagentmechanism"
	memif_mechanism "github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/memif"
)

const (
	SocketFilename = "/foo"
	BaseDir        = "baseDir"
	ID             = "testId"
)

func TestMemifServer(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	testRequest := &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Id: ID,
			Mechanism: &networkservice.Mechanism{
				Cls:  cls.LOCAL,
				Type: memif.MECHANISM,
			},
		},
	}
	suite.Run(t,
		checkvppagentmechanism.NewServerSuite(
			memif_mechanism.NewServer(BaseDir),
			memif.MECHANISM,
			func(t *testing.T, mechanism *networkservice.Mechanism) {
				m := memif.ToMechanism(mechanism)
				assert.NotNil(t, m)
				assert.Equal(t, (&url.URL{Scheme: "file", Path: filepath.Join(BaseDir, fmt.Sprintf("%s.memif.socket", ID))}).String(), m.GetSocketFileURL())
			},
			func(t *testing.T, conf *configurator.Config) {
				numInterfaces := len(conf.GetVppConfig().GetInterfaces())
				assert.Greater(t, numInterfaces, 0)
				iface := conf.GetVppConfig().GetInterfaces()[numInterfaces-1]
				assert.NotNil(t, iface)
				ifaceMemif := conf.GetVppConfig().GetInterfaces()[numInterfaces-1].GetMemif()
				assert.NotNil(t, iface)
				assert.Equal(t, filepath.Join(BaseDir, fmt.Sprintf("%s.memif.socket", ID)), ifaceMemif.GetSocketFilename())
			},
			testRequest,
			testRequest.GetConnection(),
		),
	)
}
