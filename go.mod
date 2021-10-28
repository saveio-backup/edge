module github.com/saveio/edge

go 1.16

replace (
	github.com/saveio/themis => ../themis
	github.com/saveio/dsp-go-sdk => ../dsp-go-sdk
	github.com/saveio/max => ../max
)

require (
	github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/pborman/uuid v1.2.1
	github.com/saveio/carrier v0.0.0-20210802055929-7567cc29dfc9
	github.com/saveio/dsp-go-sdk v0.0.0-20210930063117-13691f30c286
	github.com/saveio/max v0.0.0-20210825101853-a279f7982519
	github.com/saveio/pylons v0.0.0-20210802062637-12c41e6d9ba7
	github.com/saveio/scan v1.0.78
	github.com/saveio/themis v1.0.158
	github.com/saveio/themis-go-sdk v0.0.0-20210909064643-325bf09f3e3a
	github.com/tjfoc/gmtls v1.2.1 // indirect
	github.com/urfave/cli v1.22.5
)
