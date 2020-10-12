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

// Package vppagent provides a simple StartAndDialContext function that will start up a local vppagent,
// dial it, and return the grpc.ClientConnInterface
package vppagent

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"google.golang.org/grpc"

	"github.com/edwarnicke/exechelper"

	"github.com/networkservicemesh/sdk/pkg/tools/log"
)

// StartAndDialContext - starts vppagent (and vpp), dials them using any provided options and returns the resulting
// grpc.ClientConnInterface, and an error chan that will receive the error from the lifecycle of the vppagent and vpp
// and be closed when both have exited.
// Stdout and Stderr for the vppagent and vpp are set to be log.Entry(ctx).Writer().
func StartAndDialContext(ctx context.Context, opts ...Option) (vppagentCC grpc.ClientConnInterface, errCh chan error) {
	errCh = make(chan error, 5)

	o := &option{
		rootDir:  DefaultRootDir,
		grpcPort: DefaultGrpcPort,
		httpPort: DefaultHTTPPort,
	}
	for _, opt := range opts {
		opt(o)
	}

	if err := writeDefaultConfigFiles(ctx, o); err != nil {
		errCh <- err
		close(errCh)
		return nil, errCh
	}
	logWriter := log.Entry(ctx).WithField("cmd", "vpp").Writer()
	vppErrCh := exechelper.Start("vpp -c "+filepath.Join(o.rootDir, vppConfFilename),
		exechelper.WithContext(ctx),
		exechelper.WithStdout(logWriter),
		exechelper.WithStderr(logWriter),
	)
	select {
	case err := <-vppErrCh:
		errCh <- err
		close(errCh)
		return nil, errCh
	default:
	}
	logWriter = log.Entry(ctx).WithField("cmd", "vpp-agent").Writer()
	vppagentErrCh := exechelper.Start("vpp-agent -config-dir="+filepath.Join(o.rootDir, vppAgentConfDir),
		exechelper.WithContext(ctx),
		exechelper.WithStdout(logWriter),
		exechelper.WithStderr(logWriter),
	)
	select {
	case err := <-vppErrCh:
		errCh <- err
		close(errCh)
		return nil, errCh
	default:
	}
	// grpc.ClientConn for dialing the vppagent
	vppagentCC, err := grpc.DialContext(ctx, fmt.Sprintf(vppEndpoint, o.grpcPort), grpc.WithInsecure())
	if err != nil {
		errCh <- err
		close(errCh)
		return nil, errCh
	}
	go func() {
		var err error
		vppOk := true
		vppagentOk := true
		for vppOk || vppagentOk {
			select {
			case err, vppOk = <-vppErrCh:
				if vppOk {
					errCh <- err
				}
			case err, vppagentOk = <-vppagentErrCh:
				if vppagentOk {
					errCh <- err
				}
			}
		}
		close(errCh)
	}()
	return vppagentCC, errCh
}

func writeDefaultConfigFiles(ctx context.Context, o *option) error {
	configFiles := map[string]string{
		vppConfFilename:           fmt.Sprintf(vppConfContents, o.rootDir),
		vppAgentGoVPPConfFilename: fmt.Sprintf(vppAgentGoVPPConfContents, o.rootDir),
		vppAgentGrpcConfFilename:  fmt.Sprintf(vppAgentGrpcConfContents, o.grpcPort),
		vppAgentHTTPConfFilename:  fmt.Sprintf(vppAgentHTTPConfContents, o.httpPort),
		vppAgentLogsConfFilename:  vppAgentLogsConfContents,
	}
	for filename, contents := range configFiles {
		filename = filepath.Join(o.rootDir, filename)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.Entry(ctx).Infof("Configuration file: %q not found, using defaults", filename)
			if err := os.MkdirAll(path.Dir(filename), 0700); err != nil {
				return err
			}
			if err := ioutil.WriteFile(filename, []byte(contents), 0600); err != nil {
				return err
			}
		}
	}
	if err := os.MkdirAll(filepath.Join(o.rootDir, "/var/run/vpp"), 0700); os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Join(o.rootDir, "/var/log/vpp"), 0700); os.IsNotExist(err) {
		return err
	}
	return nil
}
