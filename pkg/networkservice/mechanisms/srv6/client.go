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

// Package srv6 provides networkservice chain elements that support the SRv6 Mechanism
package srv6

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/srv6"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
)

const (
	// MECHANISM string
	MECHANISM = srv6.MECHANISM
)

type srv6Client struct{}

// NewClient provides a NetworkServiceClient chain elements that support the srv6 Mechanism
func NewClient() networkservice.NetworkServiceClient {
	return &srv6Client{}
}

func (v *srv6Client) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	preferredMechanism := &networkservice.Mechanism{
		Cls:  cls.REMOTE,
		Type: srv6.MECHANISM,
	}
	request.MechanismPreferences = append(request.MechanismPreferences, preferredMechanism)
	if err := appendInterfaceConfig(ctx, request.GetConnection(), true); err != nil {
		return nil, err
	}
	return next.Client(ctx).Request(ctx, request, opts...)
}

func (v *srv6Client) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	if err := appendInterfaceConfig(ctx, conn, false); err != nil {
		return nil, err
	}
	return next.Client(ctx).Close(ctx, conn, opts...)
}
