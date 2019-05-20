#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubedb/percona/hack/gendocs
go run main.go
popd
