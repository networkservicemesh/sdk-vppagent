// Copyright (c) 2020 Doc.ai and/or its affiliates.
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

package tests

import (
	"context"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"github.com/networkservicemesh/sdk/pkg/tools/clienturlctx"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/utils/checks/testinterfaceappender"

	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/directmemif"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	memif_mechanisms "github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/memif"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/connect"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/memif"
)

const socketName = "test.sock"

func TestServerBasic(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), t.Name())
	require.Nil(t, err)
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	s := next.NewNetworkServiceServer(
		vppagent.NewServer(),
		directmemif.NewServerWithNetwork("unix"),
		memif.NewServer(dir),
		connect.NewServer(
			context.Background(),
			func(ctx context.Context, cc grpc.ClientConnInterface) networkservice.NetworkServiceClient {
				return memif.NewClient(dir)
			},
			grpc.WithInsecure(),
		),
	)
	r := request()

	_, err = s.Request(clienturlctx.WithClientURL(context.Background(), &url.URL{}), r)
	require.Nil(t, err)
	_, err = s.Close(clienturlctx.WithClientURL(context.Background(), &url.URL{}), r.Connection)
	require.Nil(t, err)
	checkThatProxyHasStopped(t, path.Join(dir, socketName))
}

func TestServerReRequest(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), t.Name())
	require.Nil(t, err)
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	s := next.NewNetworkServiceServer(
		vppagent.NewServer(),
		directmemif.NewServerWithNetwork("unix"),
		memif.NewServer(dir),
		connect.NewServer(
			context.Background(),
			func(ctx context.Context, cc grpc.ClientConnInterface) networkservice.NetworkServiceClient {
				return memif.NewClient(dir)
			},
			grpc.WithInsecure(),
		),
	)
	r := request()

	_, err = s.Request(clienturlctx.WithClientURL(context.Background(), &url.URL{}), r)
	require.Nil(t, err)
	_, err = s.Request(clienturlctx.WithClientURL(context.Background(), &url.URL{}), r)
	require.Nil(t, err)
	_, err = s.Close(clienturlctx.WithClientURL(context.Background(), &url.URL{}), r.Connection)
	require.Nil(t, err)
	checkThatProxyHasStopped(t, path.Join(dir, socketName))
}

func TestServerShouldWriteMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir, err := ioutil.TempDir(os.TempDir(), t.Name())
	require.Nil(t, err)

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	ctx = vppagent.WithConfig(ctx)
	config := vppagent.Config(ctx)
	config.VppConfig.XconnectPairs = make([]*l2.XConnectPair, 2)

	s := next.NewNetworkServiceServer(
		memif.NewServer(dir),
		connect.NewServer(
			ctx,
			func(ctx context.Context, cc grpc.ClientConnInterface) networkservice.NetworkServiceClient {
				return testinterfaceappender.NewClient()
			},
			grpc.WithInsecure(),
		),
		directmemif.NewServerWithNetwork("unix"),
	)

	r := request()
	r.Connection.Path = &networkservice.Path{
		Index: 0,
		PathSegments: []*networkservice.PathSegment{
			{
				Name: "Test",
			},
		},
	}

	conn, err := s.Request(clienturlctx.WithClientURL(ctx, &url.URL{}), r)
	require.Nil(t, err)
	metrics := conn.Path.PathSegments[conn.Path.Index].Metrics
	require.NotNil(t, metrics)
	require.Equal(t, 2, len(metrics))
	_, err = s.Close(clienturlctx.WithClientURL(ctx, &url.URL{}), r.Connection)
	require.Nil(t, err)
	checkThatProxyHasStopped(t, path.Join(dir, socketName))
}

func request() *networkservice.NetworkServiceRequest {
	return &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Id: "1",
			Mechanism: &networkservice.Mechanism{
				Cls:  cls.LOCAL,
				Type: memif_mechanisms.MECHANISM,
				Parameters: map[string]string{
					memif_mechanisms.SocketFilename: socketName,
				},
			},
		},
	}
}

func checkThatProxyHasStopped(t *testing.T, socketPath string) {
	var err error
	source, err := net.ResolveUnixAddr("unix", socketPath)
	require.Nil(t, err)
	for i := 0; i < 10; i++ {
		_, err = net.ListenUnix("unix", source)
		if err == nil {
			break
		}
		<-time.After(time.Second / 10)
	}

	require.Nil(t, err)
}
