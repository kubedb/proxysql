#!/usr/bin/env bash

pushd $GOPATH/src/kubedb.dev/proxysql/hack/gendocs
go run main.go
popd
