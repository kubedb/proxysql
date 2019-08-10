#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/kubedb.dev/percona-xtradb

source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}

IMG="proxysql-pxc"

DB_VERSION="5.7"
TAG="$DB_VERSION"

build() {
  pushd "$REPO_ROOT/hack/docker/proxysql/$DB_VERSION"

  local cmd="docker build --pull -t $DOCKER_REGISTRY/$IMG:$TAG ."
  echo $cmd
  $cmd

  popd
}

binary_repo $@
