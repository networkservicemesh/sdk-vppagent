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

// Package vppagent is a networkservice chain element that inserts a vppagent *configurator.Config into the
package vppagent

import (
	"context"

	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/ligato/vpp-agent/api/models/linux"
	"github.com/ligato/vpp-agent/api/models/netalloc"
	"github.com/ligato/vpp-agent/api/models/vpp"
)

type contextKeyType string

const (
	configKey contextKeyType = "configKey"
)

// WithConfig returns a context that contains a vppagent config
func WithConfig(ctx context.Context) context.Context {
	if config, ok := ctx.Value(configKey).(*configurator.Config); ok && config != nil {
		return ctx
	}
	rv := &configurator.Config{
		VppConfig:      &vpp.ConfigData{},
		LinuxConfig:    &linux.ConfigData{},
		NetallocConfig: &netalloc.ConfigData{},
	}
	return context.WithValue(ctx, configKey, rv)
}

// Config - returns the vppagent *configurator.Config stored in ctx
func Config(ctx context.Context) *configurator.Config {
	if rv, ok := ctx.Value(configKey).(*configurator.Config); ok {
		return rv
	}
	return nil
}
