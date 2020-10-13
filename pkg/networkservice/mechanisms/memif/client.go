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

// Package memif provides networkservice chain elements that support the memif Mechanism
package memif

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vppinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/memif"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

const (
	// MECHANISM string
	MECHANISM = memif.MECHANISM
)

type memifClient struct{}

// NewClient provides a NetworkServiceClient chain elements that support the memif Mechanism
func NewClient() networkservice.NetworkServiceClient {
	return &memifClient{}
}

func (m *memifClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	mechanism := &networkservice.Mechanism{
		Cls:        cls.LOCAL,
		Type:       memif.MECHANISM,
		Parameters: make(map[string]string),
	}
	request.MechanismPreferences = append(request.MechanismPreferences, mechanism)
	conn, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	if err := m.appendInterfaceConfig(ctx, conn); err != nil {
		return nil, err
	}
	return conn, nil
}

func (m *memifClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	rv, err := next.Client(ctx).Close(ctx, conn, opts...)
	if err != nil {
		return nil, err
	}
	if err := m.appendInterfaceConfig(ctx, conn); err != nil {
		return nil, err
	}
	return rv, nil
}

func (m *memifClient) appendInterfaceConfig(ctx context.Context, conn *networkservice.Connection) error {
	if mechanism := memif.ToMechanism(conn.GetMechanism()); mechanism != nil {
		conf := vppagent.Config(ctx)
		socketFileURL, err := url.Parse(mechanism.GetSocketFileURL())
		if err != nil {
			return errors.WithStack(err)
		}
		if socketFileURL.Scheme != "file" {
			return errors.Errorf("url scheme must be 'file' actual: %q", socketFileURL)
		}
		conf.GetVppConfig().Interfaces = append(conf.GetVppConfig().Interfaces, &vpp.Interface{
			Name:    fmt.Sprintf("client-%s", conn.GetId()),
			Type:    vppinterfaces.Interface_MEMIF,
			Enabled: true,
			Link: &vppinterfaces.Interface_Memif{
				Memif: &vppinterfaces.MemifLink{
					Master:         false,
					SocketFilename: socketFileURL.Path,
				},
			},
		})
	}
	return nil
}
