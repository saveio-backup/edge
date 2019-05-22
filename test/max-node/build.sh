#!/bin/bash
# build for fs node
if [ -f dsp ]
then
	rm dsp
fi

go build -o=dsp ../../bin/dsp/main.go
