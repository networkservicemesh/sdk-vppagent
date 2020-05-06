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
	"io/ioutil"
	"os"
	"path"

	"google.golang.org/grpc"

	"github.com/edwarnicke/exechelper"
	"github.com/networkservicemesh/sdk/pkg/tools/errctx"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
)

// StartAndDialContext - starts vppagent (and vpp), dials them using any provided options and returns the resulting
// grpc.ClientConnInterface, and an error chan that will receive the error from the lifecycle of the vppagent and vpp
// and be closed when both have exited.
// Stdout and Stderr for the vppagent and vpp are set to be log.Entry(ctx).Writer().
func StartAndDialContext(ctx context.Context, dialOptions ...grpc.DialOption) (vppagentCC grpc.ClientConnInterface, errCh chan error) {
	errCh = make(chan error, 4)
	ctx = errctx.WithErr(ctx)
	if err := writeDefaultConfigFiles(ctx); err != nil {
		errCh <- err
		close(errCh)
		return nil, errCh
	}
	logWriter := log.Entry(ctx).WithField("cmd", "vpp").Writer()
	vppErrCh := exechelper.Start("vpp -c "+vppConfFilename,
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
	vppagentErrCh := exechelper.Start("vpp-agent -config-dir="+vppAgentConfDir,
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
	vppagentCC, err := grpc.DialContext(ctx, vppEndpoint, grpc.WithInsecure())
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

func writeDefaultConfigFiles(ctx context.Context) error {
	configFiles := map[string]string{
		vppConfFilename:          vppConfContents,
		vppAgentGrpcConfFilename: vppAgentGrpcConfContents,
		vppAgentHTTPConfFilename: vppAgentHTTPConfContents,
		vppAgentLogsConfFilename: vppAgentLogsConfContents,
	}
	for filename, contents := range configFiles {
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
	return nil
}
