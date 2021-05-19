module github.com/saveio/edge

go 1.14

replace (
	github.com/saveio/carrier => ../carrier
	github.com/saveio/dsp-go-sdk => ../dsp-go-sdk
	github.com/saveio/pylons => ../pylons
	github.com/saveio/scan => ../scan
	github.com/saveio/themis-go-sdk => ../themis-go-sdk
)

require (
	github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/pborman/uuid v1.2.1
	github.com/saveio/carrier v0.0.0-20210519082359-9fc4d908c385
	github.com/saveio/dsp-go-sdk v0.0.0-00010101000000-000000000000
	github.com/saveio/max v0.0.0-20210519082655-a93c17773d75
	github.com/saveio/pylons v0.0.0-20210519083005-78a1ef20d8a0
	github.com/saveio/scan v1.0.71-0.20210519081147-e9c67b4caba0
	github.com/saveio/themis v1.0.115-0.20210519082201-29f8330c44d9
	github.com/saveio/themis-go-sdk v0.0.0-20210519082257-3f5361282350
	github.com/tjfoc/gmtls v1.2.1 // indirect
	github.com/urfave/cli v1.22.5
)
