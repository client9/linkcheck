#!/bin/sh
set -ex
go get ./...
go test .
go install .
