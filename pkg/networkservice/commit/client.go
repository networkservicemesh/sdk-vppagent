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

// Package commit provides networkservice chain elements for committing the vppagent *configurator.Config
// retrieved using vppagent.Config(ctx) to the actual vppagent instance.
package commit

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type commitClient struct {
	vppagentCC     *grpc.ClientConn
	vppagentClient configurator.ConfiguratorClient
}

// NewClient creates a NetworkServiceClient chain elements for committing the vppagent *configurator.Config
// retrieved using vppagent.Config(ctx) to the actual vppagent instance.
func NewClient(vppagentCC *grpc.ClientConn) networkservice.NetworkServiceClient {
	return &commitClient{
		vppagentCC:     vppagentCC,
		vppagentClient: configurator.NewConfiguratorClient(vppagentCC),
	}
}

func (c *commitClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	conf := vppagent.Config(ctx)
	rv, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}
	_, err = c.vppagentClient.Update(ctx, &configurator.UpdateRequest{Update: conf})
	if err != nil {
		return nil, errors.Wrapf(err, "error sending config to vppagent %s: ", conf)
	}
	return rv, nil
}

func (c *commitClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	conf := vppagent.Config(ctx)
	rv, err := next.Client(ctx).Close(ctx, conn)
	if err != nil {
		return nil, err
	}
	_, err = c.vppagentClient.Update(ctx, &configurator.UpdateRequest{Update: conf}, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "error sending config to vppagent %s: ", conf)
	}
	return rv, nil
}
