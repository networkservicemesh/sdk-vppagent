// Copyright (c) 2020 Cisco Systems, Inc.
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

// Package kernelvethpair provides networkservice chain elements that support the kernel Mechanism using veth pairs
package kernelvethpair

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/tools/kernelctx"
	"github.com/networkservicemesh/sdk-vppagent/pkg/tools/netnsinode"
)

type kernelVethPairServer struct {
	fileNameFromInodeNumberFunc func(string) (string, error)
}

// NewServer provides NetworkServiceServer chain elements that support the kernel Mechanism using veth pairs
func NewServer() networkservice.NetworkServiceServer {
	return &kernelVethPairServer{fileNameFromInodeNumberFunc: netnsinode.LinuxNetNSFileName}
}

// NewTestableServer - same as NewServer, but allows provision of fileNameFromInodeNumberFunc to allow for testing
func NewTestableServer(fileNameFromInodeNumberFunc func(string) (string, error)) networkservice.NetworkServiceServer {
	server := NewServer()
	rv := server.(*kernelVethPairServer)
	rv.fileNameFromInodeNumberFunc = fileNameFromInodeNumberFunc
	return rv
}

func (k *kernelVethPairServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	iface, err := appendInterfaceConfig(ctx, request.GetConnection(), fmt.Sprintf("server-%s", request.GetConnection().GetId()), k.fileNameFromInodeNumberFunc)
	if err != nil {
		return nil, err
	}
	ctx = kernelctx.WithServerInterface(ctx, iface)
	return next.Server(ctx).Request(ctx, request)
}

func (k *kernelVethPairServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	iface, err := appendInterfaceConfig(ctx, conn, fmt.Sprintf("server-%s", conn.GetId()), k.fileNameFromInodeNumberFunc)
	if err != nil {
		return nil, err
	}
	ctx = kernelctx.WithServerInterface(ctx, iface)
	return next.Server(ctx).Close(ctx, conn)
}
