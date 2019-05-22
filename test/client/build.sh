#!/bin/bash
# build for file uploader
if [ -f dsp ]
then
	rm dsp
fi

go build -o=dsp ../../bin/dsp/main.go
