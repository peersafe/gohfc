#!/bin/bash
set -x
cd $GOPATH/src/github.com/peersafe/gohfc/testInterface
go build  -tags "gm"
cd -