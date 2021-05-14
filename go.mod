module github.com/saveio/edge

go 1.14

replace (
	github.com/saveio/carrier => ../carrier
	github.com/saveio/dsp-go-sdk => ../dsp-go-sdk
	github.com/saveio/edge => ../edge
	github.com/saveio/max => ../max
	github.com/saveio/pylons => ../pylons
	github.com/saveio/scan => ../scan
	github.com/saveio/themis => ../themis
	github.com/saveio/themis-go-sdk => ../themis-go-sdk
)

require (
	github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/pborman/uuid v1.2.1
	github.com/saveio/carrier v0.0.0-00010101000000-000000000000
	github.com/saveio/dsp-go-sdk v0.0.0-00010101000000-000000000000
	github.com/saveio/max v0.0.0-00010101000000-000000000000
	github.com/saveio/pylons v0.0.0-00010101000000-000000000000
	github.com/saveio/scan v0.0.0-00010101000000-000000000000
	github.com/saveio/themis v0.0.0-00010101000000-000000000000
	github.com/saveio/themis-go-sdk v0.0.0-00010101000000-000000000000
	github.com/urfave/cli v1.22.5
)
