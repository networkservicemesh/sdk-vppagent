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

// Package connectioncontext provides networkservice chain elements for applying connectioncontext to the vppagent
// side of vWires being plugged into vppagent
package connectioncontext

import (
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/connectioncontext/ethernetcontext/macaddress"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/connectioncontext/ipcontext/ipaddress"
)

// NewServer creates a NetworkServiceServer chain element to set the ip address on a vpp interface
// It applies the connection context to the *vpp* side of an interface plugged into the
// Endpoint.
//                                         Endpoint
//                              +------------------------------------+
//                              |                                    |
//                              |                                    |
//                              |                                    |
//                              |                                    |
//                              |                                    |
//                              |                                    |
//                              |                                    |
//          +-------------------+ networkservice.NewServer()      |
//                              |                                    |
//                              |                                    |
//                              |                                    |
//                              |                                    |
//                              |                                    |
//                              |                                    |
//                              |                                    |
//                              +------------------------------------+
//
func NewServer() networkservice.NetworkServiceServer {
	return chain.NewNetworkServiceServer(
		ipaddress.NewServer(),
		macaddress.NewServer(),
	)
}
