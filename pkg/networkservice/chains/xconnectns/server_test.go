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

// +build !windows

package xconnectns_test

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/memif"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/endpoint"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/authorize"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"
	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/checks/checkcontext"
	"github.com/networkservicemesh/sdk/pkg/tools/grpcutils"
	"github.com/stretchr/testify/require"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/chains/xconnectns"
	memif_mechanism "github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/memif"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

const (
	anyPortLocalHost string = "127.0.0.1:0"
	sourceSocketFile string = "source.sock"
	targetSocketFile string = "target.sock"
)

type fakeVppAgent struct {
	configurator.ConfiguratorServiceServer
}

type fakeNSM struct {
}

func (fva *fakeVppAgent) Update(ctx context.Context, in *configurator.UpdateRequest) (*configurator.UpdateResponse, error) {
	return &configurator.UpdateResponse{}, nil
}

func (fn *fakeNSM) Request(ctx context.Context, req *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	req.Connection.Mechanism = &networkservice.Mechanism{
		Cls:  cls.LOCAL,
		Type: memif_mechanism.MECHANISM,
		Parameters: map[string]string{
			memif.SocketFilename: targetSocketFile,
		},
	}

	return req.Connection, nil
}

func (fn *fakeNSM) Close(context.Context, *networkservice.Connection) (*empty.Empty, error) {
	return nil, nil
}

func request(t *testing.T) *networkservice.NetworkServiceRequest {
	expires, err := ptypes.TimestampProto(time.Now().Add(time.Hour))
	require.NoError(t, err)

	return &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Mechanism: &networkservice.Mechanism{
				Cls:  cls.LOCAL,
				Type: memif_mechanism.MECHANISM,
				Parameters: map[string]string{
					memif.SocketFilename: sourceSocketFile,
				},
			},
			Path: &networkservice.Path{
				Index: 0,
				PathSegments: []*networkservice.PathSegment{
					{
						Expires: expires,
					},
				},
			},
		},
	}
}

func fakeVppClientConn(ctx context.Context, t *testing.T, css configurator.ConfiguratorServiceServer) grpc.ClientConnInterface {
	s := grpc.NewServer()
	configurator.RegisterConfiguratorServiceServer(s, css)

	vppURL := &url.URL{Scheme: "tcp", Host: anyPortLocalHost}
	grpcutils.ListenAndServe(ctx, vppURL, s)

	clientConn, err := grpc.Dial(vppURL.Host, grpc.WithInsecure())
	require.NoError(t, err)
	return clientConn
}

func tokenGenerator(peerAuthInfo credentials.AuthInfo) (tokenValue string, expireTime time.Time, err error) {
	return "TestToken", time.Date(3000, 1, 1, 1, 1, 1, 1, time.UTC), nil
}

func TestDirectMemifCase(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*2))
	defer cancel()

	dir, err := ioutil.TempDir(os.TempDir(), t.Name())
	require.NoError(t, err)

	nsmURL := &url.URL{Scheme: "tcp", Host: anyPortLocalHost}
	nsmEndpoint := endpoint.NewServer(ctx, "nsm-endpoint", authorize.NewServer(), tokenGenerator, &fakeNSM{})
	endpoint.Serve(ctx, nsmURL, nsmEndpoint)

	t.Logf("NSMgr address: %v", nsmURL.String())

	xServer := xconnectns.NewServer(ctx,
		"xconnectns-forwarder",
		authorize.NewServer(), tokenGenerator, fakeVppClientConn(ctx, t, &fakeVppAgent{}), dir,
		nil, nil, nsmURL, grpc.WithInsecure())

	server := chain.NewNetworkServiceServer(xServer, checkcontext.NewServer(t, func(t *testing.T, ctx context.Context) {
		config := vppagent.Config(ctx)
		require.NotNil(t, config)
		vppConfig := config.VppConfig
		require.NotNil(t, config)
		require.Empty(t, vppConfig.Interfaces)
		require.Empty(t, vppConfig.XconnectPairs)
	}))

	_, err = server.Request(ctx, request(t))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, sourceSocketFile))
	require.False(t, os.IsNotExist(err))
}
