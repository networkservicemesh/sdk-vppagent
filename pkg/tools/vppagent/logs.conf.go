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
	vppAgentLogsConfFilename = vppAgentConfDir + "logs.conf"
	vppAgentLogsConfContents = `# Set default config level for every plugin. Overwritten by environmental variable 'INITIAL_LOGLVL'
# Documented here: https://docs.ligato.io/en/latest/user-guide/config-files/#log-manager
default-level: warn

# Specifies a list of named loggers with respective log level
#loggers:
# - name: "agentcore",
#   level: debug
# - name: "status-check",
#   level: info
# - name: "linux-plugin",
#   level: warn
# Specifies a list of hook for logging to external links with respective
# parameters (protocol, address, port and levels) for given hook
#hooks:
#  syslog:
#    levels:
#    - panic
#    - fatal
#    - error
#    - warn
#    - info
#    - debug
#  fluent:
#    address: "10.20.30.41"
#    port: 4521
#    protocol: tcp
#    levels:
#     - error
#  logstash:
#    address: "10.20.30.42"
#    port: 123
#    protocol: tcp
`
)
