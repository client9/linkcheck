#!/bin/sh
set -ex
export GO111MODULE=on
go get ./...
go test .
go install .
