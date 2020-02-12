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
	"context"
	"io/ioutil"
	"path"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	memif_mechanism "github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/memif"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/memif"
)

const (
	SocketFilename = "foo"
	BaseDir        = "baseDir"
)

type testServer struct {
	*testing.T
	baseDir string
}

func NewTestServer(t *testing.T, baseDir string) networkservice.NetworkServiceServer {
	return &testServer{
		T:       t,
		baseDir: baseDir,
	}
}

func (t *testServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	mechanism := memif.ToMechanism(request.GetConnection().GetMechanism())
	assert.NotNil(t, mechanism)
	conf := vppagent.Config(ctx)
	assert.Greater(t, len(conf.GetVppConfig().GetInterfaces()), 0)
	assert.NotNil(t, conf.GetVppConfig().GetInterfaces()[0].GetMemif())
	assert.Equal(t, path.Join(t.baseDir, mechanism.GetSocketFilename()), conf.GetVppConfig().GetInterfaces()[0].GetMemif().GetSocketFilename())
	return next.Server(ctx).Request(ctx, request)
}

func (t *testServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	mechanism := memif.ToMechanism(conn.GetMechanism())
	assert.NotNil(t, mechanism)
	conf := vppagent.Config(ctx)
	assert.Greater(t, len(conf.GetVppConfig().GetInterfaces()), 0)
	assert.NotNil(t, conf.GetVppConfig().GetInterfaces()[0].GetMemif())
	assert.Equal(t, path.Join(t.baseDir, mechanism.GetSocketFilename()), conf.GetVppConfig().GetInterfaces()[0].GetMemif().GetSocketFilename())
	return next.Server(ctx).Close(ctx, conn)
}

func TestMemifMechanisms(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	server := chain.NewNetworkServiceServer(
		vppagent.NewServer(),
		mechanisms.NewServer(map[string]networkservice.NetworkServiceServer{
			memif.MECHANISM: memif_mechanism.NewServer(BaseDir),
		}),
		NewTestServer(t, BaseDir),
	)
	request := &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Mechanism: &networkservice.Mechanism{
				Type: memif.MECHANISM,
				Cls:  cls.LOCAL,
				Parameters: map[string]string{
					memif.SocketFilename: SocketFilename,
				},
			},
		},
	}
	conn, err := server.Request(context.Background(), request)
	assert.Nil(t, err)
	assert.NotNil(t, conn)
	_, err = server.Close(context.Background(), conn)
	assert.Nil(t, err)
}
