GOFMT=gofmt
GC=go build --tags "json1"
VERSION := $(shell git describe --abbrev=4 --always --tags)
#BUILD_DSP_PAR = -ldflags "-X github.com/saveio/edge/common/config.VERSION=$(VERSION)"
BUILD_DSP_SERVER_PAR =
BUILD_DSP_CLIENT_PAR =

SRC_FILES = $(shell git ls-files | grep -e .go$ | grep -v _test.go)

all: client

client:
	-rm ./bin/dsp/dsp
	$(GC)  $(BUILD_DSP_CLIENT_PAR) -o ./bin/dsp/dsp ./bin/dsp/main.go


do-cross: w-dsp l-dsp d-dsp

w-dsp:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GC) $(BUILD_DSP_SERVER_PAR) -o do-windows-amd64-ddns.exe ./bin/ddns/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GC) $(BUILD_DSP_CLIENT_PAR) -o do-windows-amd64-dsp.exe ./bin/dsp/main.go

l-dsp:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_DSP_SERVER_PAR) -o do-linux-amd64-ddns ./bin/ddns/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_DSP_CLIENT_PAR) -o do-linux-amd64-dsp ./bin/dsp/main.go

d-dsp:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_DSP_SERVER_PAR) -o do-darwin-amd64-ddns ./bin/ddns/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_DSP_CLIENT_PAR) -o do-darwin-amd64-dsp ./bin/dsp/main.go

format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6 *exe
