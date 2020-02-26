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

package ipaddress_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"
	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/inject/injecterror"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/connectioncontextkernel/ipcontext/ipaddress"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/kernel/kerneltap"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

func serverRequest() *networkservice.NetworkServiceRequest {
	return &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Mechanism: &networkservice.Mechanism{
				Type: kernel.MECHANISM,
			},
			Context: &networkservice.ConnectionContext{
				IpContext: &networkservice.IPContext{
					DstIpAddr: IPAddress,
				},
			},
		},
	}
}

func TestSetIPKernelServer_Request(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	server := chain.NewNetworkServiceServer(
		kerneltap.NewTestableServer(fileInodeToFilename),
		ipaddress.NewServer(),
	)
	ctx := vppagent.WithConfig(context.Background())
	conn, err := server.Request(ctx, serverRequest())

	assert.NotNil(t, conn)
	assert.Nil(t, err)

	conf := vppagent.Config(ctx)
	numInterfaces := len(conf.GetLinuxConfig().GetInterfaces())
	require.Greater(t, numInterfaces, 0)
	ipAddresses := conf.GetLinuxConfig().GetInterfaces()[numInterfaces-1].GetIpAddresses()
	require.Greater(t, len(ipAddresses), 0)
	assert.Equal(t, IPAddress, ipAddresses[0])
}

func TestSetIPKernelServer_Close(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	server := chain.NewNetworkServiceServer(
		kerneltap.NewTestableServer(fileInodeToFilename),
		ipaddress.NewServer(),
	)
	ctx := vppagent.WithConfig(context.Background())
	_, err := server.Close(ctx, serverRequest().GetConnection())

	assert.Nil(t, err)

	conf := vppagent.Config(ctx)
	numInterfaces := len(conf.GetLinuxConfig().GetInterfaces())
	require.Greater(t, numInterfaces, 0)
	ipAddresses := conf.GetLinuxConfig().GetInterfaces()[numInterfaces-1].GetIpAddresses()
	require.Greater(t, len(ipAddresses), 0)
	assert.Equal(t, IPAddress, ipAddresses[0])
}

func TestSetIPKernelServerPropagatesError(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	server := chain.NewNetworkServiceServer(
		ipaddress.NewServer(),
		injecterror.NewServer(),
	)
	_, err := server.Request(vppagent.WithConfig(context.Background()), serverRequest())
	assert.NotNil(t, err)
	_, err = server.Close(vppagent.WithConfig(context.Background()), serverRequest().GetConnection())
	assert.NotNil(t, err)
}
