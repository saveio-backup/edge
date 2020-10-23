GOFMT=gofmt
GC=go build --tags "json1"

EDGE_GITCOMMIT=$(shell git log -1 --pretty=format:"%H")
PYLONS_GITCOMMIT=$(shell cd .. && cd pylons && git log -1 --pretty=format:"%H")
CARRIER_GITCOMMIT=$(shell cd .. && cd carrier && git log -1 --pretty=format:"%H")
MAX_GITCOMMIT=$(shell cd .. && cd max && git log -1 --pretty=format:"%H")
DSP_GITCOMMIT=$(shell cd .. && cd dsp-go-sdk && git log -1 --pretty=format:"%H")
SCAN_GITCOMMIT=$(shell cd .. && cd scan && git log -1 --pretty=format:"%H")
ARM64_CC=/home/dasein/toolchain-aarch64_cortex-a53_gcc-8.2.0_glibc/bin/aarch64-openwrt-linux-gnu-gcc
BUILD_EDGE_PAR =-v -ldflags "-s -w -X github.com/saveio/edge/dsp.Version=$(EDGE_GITCOMMIT) -X github.com/saveio/pylons.Version=${PYLONS_GITCOMMIT} -X github.com/saveio/carrier/network.Version=${CARRIER_GITCOMMIT} -X github.com/saveio/max/max.Version=${MAX_GITCOMMIT} -X github.com/saveio/dsp-go-sdk/dsp.Version=${DSP_GITCOMMIT} -X github.com/saveio/scan/common/config.VERSION=$(SCAN_GITCOMMIT)"

all: client

client:
	$(GC)  $(BUILD_EDGE_PAR) -o ./edge ./bin/edge/main.go


do-cross: w-dsp l-dsp d-dsp

w-dsp:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 $(GC) $(BUILD_EDGE_PAR) -o edge-windows-amd64.exe ./bin/edge/main.go

l-dsp:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_EDGE_PAR) -o edge-linux-amd64 ./bin/edge/main.go

d-dsp:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_EDGE_PAR) -o edge-darwin-amd64 ./bin/edge/main.go
arm64-dsp:
       CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=${ARM64_CC} $(GC) $(BUILD_EDGE_PAR) -o edge-linux-arm64 ./bin/edge/main.go
format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6 *exe