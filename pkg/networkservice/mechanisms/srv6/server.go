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

package srv6

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

type srv6Server struct{}

// NewServer provides a NetworkServiceServer chain elements that support the srv6 Mechanism
func NewServer() networkservice.NetworkServiceServer {
	return &srv6Server{}
}

func (v *srv6Server) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	if err := appendInterfaceConfig(ctx, request.GetConnection(), true); err != nil {
		return nil, err
	}
	return next.Server(ctx).Request(ctx, request)
}

func (v *srv6Server) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	if err := appendInterfaceConfig(ctx, conn, false); err != nil {
		return nil, err
	}
	return next.Server(ctx).Close(ctx, conn)
}
