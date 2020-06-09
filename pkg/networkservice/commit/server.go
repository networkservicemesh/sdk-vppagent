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

package commit

import (
	"context"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

type commitServer struct {
	vppagentCC     grpc.ClientConnInterface
	vppagentClient configurator.ConfiguratorServiceClient
	sync.Once
}

// NewServer creates a NetworkServiceServer chain elements for committing the vppagent *configurator.Config
// retrieved using vppagent.Config(ctx) to the actual vppagent instance.
func NewServer(vppagentCC grpc.ClientConnInterface) networkservice.NetworkServiceServer {
	return &commitServer{
		vppagentCC:     vppagentCC,
		vppagentClient: configurator.NewConfiguratorServiceClient(vppagentCC),
	}
}

func (c *commitServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	// First time we connect we need to do a FullResync
	var fullResync bool
	c.Do(func() {
		fullResync = true
	})
	conf := vppagent.Config(ctx)
	_, err := c.vppagentClient.Update(ctx, &configurator.UpdateRequest{Update: conf, FullResync: fullResync}, grpc.WaitForReady(true))
	if err != nil {
		return nil, errors.Wrapf(err, "error sending config to vppagent %s: ", conf)
	}
	return next.Server(ctx).Request(ctx, request)
}

func (c *commitServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	conf := vppagent.Config(ctx)
	_, err := c.vppagentClient.Delete(ctx, &configurator.DeleteRequest{Delete: conf})
	if err != nil {
		return nil, err
	}
	return next.Server(ctx).Close(ctx, conn)
}
