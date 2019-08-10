#!/usr/bin/env bash

pushd $GOPATH/src/kubedb.dev/percona-xtradb/hack/gendocs
go run main.go
popd
