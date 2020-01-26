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

package connectioncontext

import (
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"

	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/connectioncontext/ethernetcontext/macaddress"
	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/connectioncontext/ipcontext/ipaddress"
)

// NewClient creates a NetworkServiceClient chain element to set the ip address on a vpp interface
// It applies the connection context to the *vpp* side of an interface leaving the
// Endpoint.
//                                               Endpoint
//                              +-------------------------------------+
//                              |                                     |
//                              |                                     |
//                              |                                     |
//                              |                                     |
//                              |                                     |
//                              |                                     |
//                              |                                     |
//                              |        connectioncontext.NewClient()+-------------------+
//                              |                                     |
//                              |                                     |
//                              |                                     |
//                              |                                     |
//                              |                                     |
//                              |                                     |
//                              |                                     |
//                              +-------------------------------------+
//
func NewClient() networkservice.NetworkServiceClient {
	return chain.NewNetworkServiceClient(
		ipaddress.NewClient(),
		macaddress.NewClient(),
	)
}
