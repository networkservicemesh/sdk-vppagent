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

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/networkservicemesh/sdk/pkg/tools/errctx"
	"github.com/networkservicemesh/sdk/pkg/tools/executils"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
)

// StartAndDialContext - starts vppagent (and vpp), dials them using any provided options and returns the resulting
// grpc.ClientConnInterface, a context for the running vppagent (and vpp) servers and any error that occurred
// The returned context will be canceled with the vpp or vppagent servers die.  In addition, you can retrieve the error
// for their death from that context using errctx.Err(...).
// Stdout and Stderr for the vppagent and vpp are set to be log.Entry(ctx).Writer().
func StartAndDialContext(ctx context.Context, dialOptions ...grpc.DialOption) (grpc.ClientConnInterface, context.Context, error) {
	ctx, cancel := context.WithCancel(ctx)
	ctx = errctx.WithErr(ctx)
	if err := writeDefaultConfigFiles(ctx); err != nil {
		errctx.SetErr(ctx, err)
		cancel()
		return nil, ctx, err
	}
	vppCtx, err := executils.Start(log.WithField(ctx, "cmd", "vpp"), "vpp -c "+vppConfFilename)
	if err != nil {
		errctx.SetErr(ctx, err)
		cancel()
		return nil, ctx, err
	}
	vppagentCtx, err := executils.Start(log.WithField(ctx, "cmd", "vpp-agent"), "vpp-agent -config-dir="+vppAgentConfDir)
	if err != nil {
		errctx.SetErr(ctx, err)
		cancel()
		return nil, ctx, err
	}
	// grpc.ClientConn for dialing the vppagent
	vppagentCC, err := grpc.DialContext(ctx, vppEndpoint, grpc.WithInsecure())
	if err != nil {
		errctx.SetErr(ctx, err)
		cancel()
		return nil, ctx, err
	}
	go func() {
		select {
		case <-vppCtx.Done():
			errctx.SetErr(ctx, errors.Errorf("vpp quit unexpectedly: %+v", errctx.Err(vppCtx)))
			cancel()
		case <-vppagentCtx.Done():
			errctx.SetErr(ctx, errors.Errorf("vpp-agent quit unexpectedly: %+v", errctx.Err(vppagentCtx)))
			cancel()
		case <-ctx.Done():
		}
	}()
	return vppagentCC, ctx, nil
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
