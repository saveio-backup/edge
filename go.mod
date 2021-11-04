module github.com/saveio/edge

go 1.16

replace (
	github.com/saveio/dsp-go-sdk => ../dsp-go-sdk
	github.com/saveio/max => ../max
	github.com/saveio/themis => ../themis
)

require (
	github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/pborman/uuid v1.2.1
	github.com/saveio/carrier v0.0.0-20210802055929-7567cc29dfc9
	github.com/saveio/dsp-go-sdk v0.0.0-20210930063117-13691f30c286
	github.com/saveio/max v0.0.0-20211028065147-9634b553b277
	github.com/saveio/pylons v0.0.0-20211101063731-21f49b2a554f
	github.com/saveio/scan v1.0.78
	github.com/saveio/themis v1.0.161
	github.com/saveio/themis-go-sdk v0.0.0-20211028090828-c8f7c2674776
	github.com/tjfoc/gmtls v1.2.1 // indirect
	github.com/urfave/cli v1.22.5
)
