module github.com/enobufs/go-rudp

go 1.12

require (
	github.com/pion/logging v0.2.2
	github.com/pion/sctp v1.7.0
	github.com/stretchr/testify v1.4.0
	gopkg.in/yaml.v2 v2.2.4 // indirect
)

replace github.com/pion/sctp => ../../pion/sctp
