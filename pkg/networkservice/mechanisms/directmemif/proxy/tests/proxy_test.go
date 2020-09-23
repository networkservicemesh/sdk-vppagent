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
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/mechanisms/directmemif/proxy"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	sourceSocket = "source.sock"
	targetSocket = "target.sock"
)

func TestClosingOpeningMemifProxy(t *testing.T) {
	p, err := proxy.New(sourceSocket, targetSocket, "unix", nil)
	require.Nil(t, err)
	for i := 0; i < 10; i++ {
		err = p.Start()
		require.Nil(t, err)
		err = p.Stop()
		require.Nil(t, err)
	}
}

func TestTransferBetweenMemifProxies(t *testing.T) {
	for i := 0; i < 10; i++ {
		p1, err := proxy.New(sourceSocket, targetSocket, "unix", nil)
		require.Nil(t, err)
		p2, err := proxy.New(targetSocket, sourceSocket, "unix", nil)
		require.Nil(t, err)
		err = p1.Start()
		require.Nil(t, err)
		err = p2.Start()
		require.Nil(t, err)
		err = connectAndSendMsg(sourceSocket)
		require.Nil(t, err)
		err = connectAndSendMsg(targetSocket)
		require.Nil(t, err)
		err = p1.Stop()
		require.Nil(t, err)
		err = p2.Stop()
		require.Nil(t, err)
	}
}

func TestProxyListenerCalled(t *testing.T) {
	proxyState := uint32(0)
	p, err := proxy.New(sourceSocket, targetSocket, "unix", proxy.StopListenerAdapter(func() {
		atomic.StoreUint32(&proxyState, 1)
	}))
	require.Nil(t, err)
	err = p.Start()
	require.Nil(t, err)
	err = p.Stop()
	require.Nil(t, err)
	for t := time.Now(); time.Since(t) < time.Second; {
		if atomic.LoadUint32(&proxyState) != 0 {
			break
		}
	}
	require.True(t, atomic.LoadUint32(&proxyState) != 0)
}

func TestProxyListenerCalledOnDestroySocketFile(t *testing.T) {
	proxyState := uint32(0)
	p, err := proxy.New(sourceSocket, targetSocket, "unix", proxy.StopListenerAdapter(func() {
		atomic.StoreUint32(&proxyState, 1)
	}))
	require.Nil(t, err)
	err = p.Start()
	require.Nil(t, err)
	err = connectAndSendMsg(sourceSocket)
	require.Nil(t, err)
	err = os.Remove(sourceSocket)
	require.Nil(t, err)
	for t := time.Now(); time.Since(t) < time.Second; {
		if atomic.LoadUint32(&proxyState) != 0 {
			break
		}
	}
	require.True(t, atomic.LoadUint32(&proxyState) != 0)
}

func TestStartProxyIfSocketFileIsExist(t *testing.T) {
	_, err := os.Create(sourceSocket)
	require.Nil(t, err)
	p, err := proxy.New(sourceSocket, targetSocket, "unix", nil)
	require.Nil(t, err)
	err = p.Start()
	require.Nil(t, err)
	err = p.Stop()
	require.Nil(t, err)
}

func TestUpdateMetrics(t *testing.T) {
	p1, err := proxy.New(sourceSocket, targetSocket, "unix", nil)
	require.Nil(t, err)
	p2, err := proxy.New(targetSocket, sourceSocket, "unix", nil)
	require.Nil(t, err)
	err = p1.Start()
	require.Nil(t, err)
	err = p2.Start()
	require.Nil(t, err)
	err = connectAndSendMsg(sourceSocket)
	require.Nil(t, err)
	require.Eventually(t, func() bool {
		return len(p1.Metrics()) == 2 && p1.Metrics()["tx_bytes"] == "6"
	}, time.Millisecond*200, time.Millisecond*50)
	err = p1.Stop()
	require.Nil(t, err)
	err = p2.Stop()
	require.Nil(t, err)
	t.Log(p1.Metrics())
}

func connectAndSendMsg(sock string) error {
	addr, err := net.ResolveUnixAddr("unix", sock)
	if err != nil {
		return err
	}

	var conn *net.UnixConn
	conn, err = net.DialUnix("unix", nil, addr)
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte("secret"))
	if err != nil {
		return err
	}

	err = conn.Close()
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	return nil
}
