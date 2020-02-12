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
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/memif"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	memif_mechanism "github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/memif"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type beforeClient struct {
	*testing.T
	baseDir string
}

func NewBeforeClient(t *testing.T, baseDir string) networkservice.NetworkServiceClient {
	return &beforeClient{
		T:       t,
		baseDir: baseDir,
	}
}

func (b beforeClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	conn, err := next.Client(ctx).Request(ctx, request)
	assert.NotNil(b, conn)
	assert.Nil(b, err)
	mechanism := memif.ToMechanism(conn.GetMechanism())
	assert.NotNil(b, mechanism)
	conf := vppagent.Config(ctx)
	assert.Greater(b, len(conf.GetVppConfig().GetInterfaces()), 0)
	numInterfaces := len(conf.GetVppConfig().GetInterfaces())
	iface := conf.GetVppConfig().GetInterfaces()[numInterfaces-1].GetMemif()
	assert.NotNil(b, iface)
	assert.Equal(b, path.Join(b.baseDir, mechanism.GetSocketFilename()), iface.GetSocketFilename())
	return conn, err
}

func (b beforeClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	_, err := next.Client(ctx).Close(ctx, conn)
	assert.Nil(b, err)
	mechanism := memif.ToMechanism(conn.GetMechanism())
	assert.NotNil(b, mechanism)
	conf := vppagent.Config(ctx)
	assert.Greater(b, len(conf.GetVppConfig().GetInterfaces()), 0)
	numInterfaces := len(conf.GetVppConfig().GetInterfaces())
	iface := conf.GetVppConfig().GetInterfaces()[numInterfaces-1].GetMemif()
	assert.NotNil(b, iface)
	assert.Equal(b, path.Join(b.baseDir, mechanism.GetSocketFilename()), iface.GetSocketFilename())
	return &empty.Empty{}, err
}

type afterClient struct {
	*testing.T
}

func NewAfterClient(t *testing.T) networkservice.NetworkServiceClient {
	return &afterClient{t}
}

func (a *afterClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	var found bool
	for _, mechanism := range request.GetMechanismPreferences() {
		if m := memif.ToMechanism(mechanism); m != nil {
			found = true
			request.GetConnection().Mechanism = mechanism
			mechanism.Parameters[memif.SocketFilename] = SocketFilename
		}
	}
	assert.True(a, found, "Did not find memif Mechanism in MechanismPreferences")
	return next.Client(ctx).Request(ctx, request)
}

func (a *afterClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	return next.Client(ctx).Close(ctx, conn)
}

// TestMemifClient - tests to make sure that the memif mechanism client does its two jobs:
//        - Add memif to the Mechanisms preferences before it calls next.Client(ctx).Request(...)
//        - Add the proper Interface to vppagent.Config() after it gets back the fully completed after calling
//            next.Client(ctx).Request(...)
func TestMemifClient(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	client := chain.NewNetworkServiceClient(
		vppagent.NewClient(),
		NewBeforeClient(t, BaseDir),
		memif_mechanism.NewClient(BaseDir),
		NewAfterClient(t),
	)
	request := &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{},
	}
	conn, err := client.Request(context.Background(), request)
	assert.NotNil(t, conn)
	assert.Nil(t, err)
	_, err = client.Close(context.Background(), conn)
	assert.Nil(t, err)
}
