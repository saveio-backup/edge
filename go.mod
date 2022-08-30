module github.com/saveio/edge

go 1.16

require (
	github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/pborman/uuid v1.2.1
	github.com/saveio/carrier v0.0.0-20210802055929-7567cc29dfc9
	github.com/saveio/dsp-go-sdk v0.0.0-20220803090127-d0b9f0ed40e3
	github.com/saveio/max v0.0.0-20220721095517-c20d7e72d0aa
	github.com/saveio/pylons v0.0.0-20220209062224-f4c541f85b18
	github.com/saveio/scan v1.0.97-0.20220331091913-ea1ec535c921
	github.com/saveio/themis v1.0.171
	github.com/saveio/themis-go-sdk v0.0.0-20220808082100-56c56fa0b0b2
	github.com/tjfoc/gmtls v1.2.1 // indirect
	github.com/urfave/cli v1.22.5
)

replace (
	github.com/saveio/carrier => ../carrier
	github.com/saveio/dsp-go-sdk => ../dsp-go-sdk
	github.com/saveio/max => ../max
	github.com/saveio/pylons => ../pylons
	github.com/saveio/scan => ../scan
	github.com/saveio/themis => ../themis
	github.com/saveio/themis-go-sdk => ../themis-go-sdk
	github.com/tjfoc/gmsm => github.com/tjfoc/gmsm v1.3.1
)
