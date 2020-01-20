// Copyright (c) 2018-2020 VMware, Inc.
//
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

// Package acl provides a NetworkServiceServer chain element to apply an ingress acl
package acl

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ligato/vpp-agent/api/configurator"
	vppacl "github.com/ligato/vpp-agent/api/models/vpp/acl"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/api/networkservice"

	"github.com/networkservicemesh/sdk-vppagent/pkg/networkservice/vppagent"
)

// ACL is a VPP Agent ACL composite
type acl struct {
	rules []*vppacl.ACL_Rule
}

// NewServer creates a NetworkServiceServer that applies an ingress acl specified by rules
func NewServer(rules []*vppacl.ACL_Rule) networkservice.NetworkServiceServer {
	return &acl{rules: rules}
}

func (a *acl) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*connection.Connection, error) {
	conf := vppagent.Config(ctx)
	a.appendACLConfig(conf)
	return next.Server(ctx).Request(ctx, request)
}

func (a *acl) Close(ctx context.Context, conn *connection.Connection) (*empty.Empty, error) {
	conf := vppagent.Config(ctx)
	a.appendACLConfig(conf)
	return next.Server(ctx).Close(ctx, conn)
}

func (a *acl) appendACLConfig(conf *configurator.Config) {
	if a.rules != nil && len(conf.GetVppConfig().GetInterfaces()) > 0 {
		// TODO - this can likely be changed into just a single ACL, with appending new interface to which it
		// can be applied
		conf.GetVppConfig().Acls = append(conf.GetVppConfig().Acls, &vppacl.ACL{
			Name:  "ingress-acl-" + conf.GetVppConfig().GetInterfaces()[len(conf.GetVppConfig().GetInterfaces())-1].GetName(),
			Rules: a.rules,
			Interfaces: &vppacl.ACL_Interfaces{
				Egress:  []string{},
				Ingress: []string{conf.GetVppConfig().GetInterfaces()[len(conf.GetVppConfig().GetInterfaces())-1].GetName()},
			},
		})
	}
}
