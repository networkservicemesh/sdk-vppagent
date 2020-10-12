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

package vppagent

const (
	// DefaultRootDir - Default value for RootDir
	DefaultRootDir = ""
	// DefaultGrpcPort - Default value for GRPC port
	DefaultGrpcPort = 9111
	// DefaultHTTPPort - Default value for HTTP port
	DefaultHTTPPort = 9191
)

type option struct {
	rootDir  string
	grpcPort int
	httpPort int
}

// Option - Option for use with vppagent.Start(...)
type Option func(opt *option)

// WithRootDir - set a root dir (usually a tmpDir) for all conf files.
func WithRootDir(rootDir string) Option {
	return func(opt *option) {
		opt.rootDir = rootDir
	}
}

// WithGrpcPort - set the grpc port for vppagent to listen on
func WithGrpcPort(grpcPort int) Option {
	return func(opt *option) {
		opt.grpcPort = grpcPort
	}
}

// WithHTTPPort - set the http port for vppagent to listen on
func WithHTTPPort(httpPort int) Option {
	return func(opt *option) {
		opt.httpPort = httpPort
	}
}
