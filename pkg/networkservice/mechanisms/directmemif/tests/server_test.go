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

	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/directmemif"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/clienturl"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	memif_mechanisms "github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/memif"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/connect"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/memif"
)

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
	r := &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Id: "1",
			Mechanism: &networkservice.Mechanism{
				Cls:  cls.LOCAL,
				Type: memif_mechanisms.MECHANISM,
				Parameters: map[string]string{
					memif_mechanisms.SocketFilename: "test.sock",
				},
			},
		},
	}
	_, err = s.Request(clienturl.WithClientURL(context.Background(), &url.URL{}), r)
	require.Nil(t, err)
	_, err = s.Close(clienturl.WithClientURL(context.Background(), &url.URL{}), r.Connection)
	require.Nil(t, err)
	checkThatProxyHasStopped(t, path.Join(dir, "test.sock"))
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
	r := &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Id: "1",
			Mechanism: &networkservice.Mechanism{
				Cls:  cls.LOCAL,
				Type: memif_mechanisms.MECHANISM,
				Parameters: map[string]string{
					memif_mechanisms.SocketFilename: "test.sock",
				},
			},
		},
	}
	_, err = s.Request(clienturl.WithClientURL(context.Background(), &url.URL{}), r)
	require.Nil(t, err)
	_, err = s.Request(clienturl.WithClientURL(context.Background(), &url.URL{}), r)
	require.Nil(t, err)
	_, err = s.Close(clienturl.WithClientURL(context.Background(), &url.URL{}), r.Connection)
	require.Nil(t, err)
	checkThatProxyHasStopped(t, path.Join(dir, "test.sock"))
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
