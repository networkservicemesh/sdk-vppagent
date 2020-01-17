module github.com/networkservicemesh/sdk-vppagent

go 1.13

require (
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/kr/pretty v0.2.0 // indirect
	github.com/ligato/cn-infra v2.2.0+incompatible // indirect
	github.com/ligato/vpp-agent v2.5.1+incompatible
	github.com/networkservicemesh/networkservicemesh/controlplane/api v0.2.0
	github.com/networkservicemesh/sdk v0.0.0-20200116174906-b470c807ed55
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace github.com/networkservicemesh/networkservicemesh/controlplane/api => github.com/networkservicemesh/networkservicemesh/controlplane/api v0.0.0-20200116214944-fa2a5cf363dd

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
