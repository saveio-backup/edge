GOFMT=gofmt
GC=go build --tags "json1"
VERSION := $(shell git describe --abbrev=4 --always --tags)
BUILD_EDGE_PAR =-x -v -ldflags "-s -w -X github.com/saveio/edge/dsp.GitCommit=$(GITCOMMIT)"

SRC_FILES = $(shell git ls-files | grep -e .go$ | grep -v _test.go)

all: client

client:
	$(GC)  $(BUILD_DSP_CLIENT_PAR) -o ./edge ./bin/edge/main.go


do-cross: w-dsp l-dsp d-dsp

w-dsp:
	$(eval GITCOMMIT=$(shell git log -1 --pretty=format:"%H"))
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 $(GC) $(BUILD_EDGE_PAR) -o edge-windows-amd64.exe ./bin/edge/main.go

l-dsp:
	$(eval GITCOMMIT=$(shell git log -1 --pretty=format:"%H"))
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_EDGE_PAR) -o edge-linux-amd64 ./bin/edge/main.go

d-dsp:
	$(eval GITCOMMIT=$(shell git log -1 --pretty=format:"%H"))
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_EDGE_PAR) -o edge-darwin-amd64 ./bin/edge/main.go

format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6 *exe
