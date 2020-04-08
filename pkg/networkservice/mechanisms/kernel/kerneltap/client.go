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

package kerneltap

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"

	"github.com/networkservicemesh/sdk-vppagent/pkg/tools/netnsinode"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type kernelTapClient struct {
	fileNameFromInodeNumberFunc func(string) (string, error)
}

// NewClient provides NetworkServiceClient chain elements that support the kernel Mechanism using tapv2
func NewClient() networkservice.NetworkServiceClient {
	return &kernelTapClient{
		fileNameFromInodeNumberFunc: netnsinode.LinuxNetNSFileName,
	}
}

// NewTestableClient - same as NewClient, but allows provision of fileNameFromInodeNumberFunc to allow for testing
func NewTestableClient(fileNameFromInodeNumberFunc func(string) (string, error)) networkservice.NetworkServiceClient {
	client := NewClient()
	rv := client.(*kernelTapClient)
	rv.fileNameFromInodeNumberFunc = fileNameFromInodeNumberFunc
	return rv
}

func (k *kernelTapClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	mechanism := &networkservice.Mechanism{
		Cls:  cls.LOCAL,
		Type: kernel.MECHANISM,
	}
	request.MechanismPreferences = append(request.MechanismPreferences, mechanism)
	conn, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	if err := appendInterfaceConfig(ctx, conn, fmt.Sprintf("client-%s", conn.GetId()), k.fileNameFromInodeNumberFunc); err != nil {
		return nil, err
	}
	return conn, nil
}

func (k *kernelTapClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	rv, err := next.Client(ctx).Close(ctx, conn, opts...)
	if configErr := appendInterfaceConfig(ctx, conn, fmt.Sprintf("client-%s", conn.GetId()), k.fileNameFromInodeNumberFunc); configErr != nil {
		return nil, configErr
	}
	return rv, err
}
