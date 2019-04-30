#!/bin/bash
# build for client
if [ -f dsp ]
then
	rm dsp
fi

go build -o=dsp ../../bin/dsp/main.go
