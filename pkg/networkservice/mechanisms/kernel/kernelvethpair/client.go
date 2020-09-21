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

// +build !windows

package kernelvethpair

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type kernelVethPairClient struct{}

// NewClient provides NetworkServiceClient chain elements that support the kernel Mechanism using veth pairs
func NewClient() networkservice.NetworkServiceClient {
	return &kernelVethPairClient{}
}

func (k *kernelVethPairClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	preferredMechanism := &networkservice.Mechanism{
		Cls:  cls.LOCAL,
		Type: kernel.MECHANISM,
	}
	request.MechanismPreferences = append(request.MechanismPreferences, preferredMechanism)
	conn, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	if err := k.appendInterfaceConfig(ctx, conn); err != nil {
		return nil, err
	}
	return conn, nil
}

func (k *kernelVethPairClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	rv, err := next.Client(ctx).Close(ctx, conn, opts...)
	if err != nil {
		return nil, err
	}
	err = k.appendInterfaceConfig(ctx, conn)
	if err != nil {
		return nil, err
	}
	return rv, err
}

func (k *kernelVethPairClient) appendInterfaceConfig(ctx context.Context, conn *networkservice.Connection) error {
	if mechanism := kernel.ToMechanism(conn.GetMechanism()); mechanism != nil {
		netNSURLStr := mechanism.GetNetNSURL()
		netNSURL, err := url.Parse(netNSURLStr)
		if err != nil {
			return err
		}
		if netNSURL.Scheme != fileScheme {
			return errors.Errorf("kernel.ToMechanism(conn.GetMechanism()).GetNetNSURL() must be of scheme %q: %q", fileScheme, netNSURL)
		}
		appendInterfaceConfig(vppagent.Config(ctx), fmt.Sprintf("client-%s", conn.GetId()), kernel.ToMechanism(conn.GetMechanism()).GetInterfaceName(conn), netNSURL.Path)
	}
	return nil
}
