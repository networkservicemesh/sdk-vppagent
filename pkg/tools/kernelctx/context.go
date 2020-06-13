// Copyright (c) 2020 Cisco and/or its affiliates.
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

// Package kernelctx enables the server side kernel interface to be stored in the context
package kernelctx

import (
	"context"

	linuxinterfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
)

type contextKeyType string

const (
	serverInterfaceKey contextKeyType = "serverInterface"
)

// WithServerInterface - Server interface iface to the context
func WithServerInterface(ctx context.Context, iface *linuxinterfaces.Interface) context.Context {
	return context.WithValue(ctx, serverInterfaceKey, iface)
}

// ServerInterface - retrieve server interface from the context
func ServerInterface(ctx context.Context) *linuxinterfaces.Interface {
	rvRaw := ctx.Value(serverInterfaceKey)
	if rvRaw != nil {
		return rvRaw.(*linuxinterfaces.Interface)
	}
	return nil
}
