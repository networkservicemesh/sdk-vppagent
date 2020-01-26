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

package connectioncontextkernel

import (
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/connectioncontextkernel/ethernetcontext/macaddress"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/connectioncontextkernel/ipcontext/ipaddress"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/connectioncontextkernel/ipcontext/routes"
)

// NewServer provides a NetworkServiceServer that applies the connection context to a kernel interface
// It applies the connection context on the *kernel* side of an interface plugged into the
// Endpoint.  Generally only used by privileged Endpoints like those implementing
// the Cross Connect Network Service for K8s (formerly known as NSM Forwarder).
//                                                      Endpoint
//                                            +---------------------------+
//                                            |                           |
//                                            |                           |
//                                            |                           |
//                                            |                           |
//                                            |                           |
//                                            |                           |
//                                            |                           |
//                        +-------------------+                           |
//  connectioncontextkernel.NewServer()       |                           |
//                                            |                           |
//                                            |                           |
//                                            |                           |
//                                            |                           |
//                                            |                           |
//                                            |                           |
//                                            +---------------------------+
//
func NewServer() networkservice.NetworkServiceServer {
	return chain.NewNetworkServiceServer(
		ipaddress.NewServer(),
		macaddress.NewServer(),
		// Note: routes are only applicable in this circumstance in the server side
		routes.NewServer(),
	)
}
