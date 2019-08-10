#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=${GOPATH}/src/kubedb.dev/percona-xtradb

export DB_UPDATE=0
export TOOLS_UPDATE=0
export EXPORTER_UPDATE=0
export OPERATOR_UPDATE=0
export PROXYSQL_UPDATE=0

show_help() {
  echo "update-docker.sh [options]"
  echo " "
  echo "options:"
  echo "-h, --help                       show brief help"
  echo "    --db-only                    update only database images"
  echo "    --tools-only                 update only database-tools images"
  echo "    --exporter-only              update only database-exporter images"
  echo "    --operator-only              update only operator image"
  echo "    --proxysql-only              update only proxysql image"
}

while test $# -gt 0; do
  case "$1" in
    -h | --help)
      show_help
      exit 0
      ;;
    --db-only)
      export DB_UPDATE=1
      shift
      ;;
    --tools-only)
      export TOOLS_UPDATE=1
      shift
      ;;
    --exporter-only)
      export EXPORTER_UPDATE=1
      shift
      ;;
    --operator-only)
      export OPERATOR_UPDATE=1
      shift
      ;;
    --proxysql-only)
      export PROXYSQL_UPDATE=1
      shift
      ;;
    *)
      show_help
      exit 1
      ;;
  esac
done

dbversions=(
  5.7
)

exporters=(
  v0.11.0
)

echo ""
env | sort | grep -e DOCKER_REGISTRY -e APPSCODE_ENV || true
echo ""

if [ "$DB_UPDATE" -eq 1 ]; then
  cowsay -f tux "Processing database images" || true
  for db in "${dbversions[@]}"; do
    ${REPO_ROOT}/hack/docker/percona/${db}/make.sh build
    ${REPO_ROOT}/hack/docker/percona/${db}/make.sh push
  done
fi

if [ "$EXPORTER_UPDATE" -eq 1 ]; then
  cowsay -f tux "Processing database-exporter images" || true
  for exporter in "${exporters[@]}"; do
    ${REPO_ROOT}/hack/docker/mysqld-exporter/${exporter}/make.sh
  done
fi

if [ "$OPERATOR_UPDATE" -eq 1 ]; then
  cowsay -f tux "Processing Operator images" || true
  ${REPO_ROOT}/hack/docker/percona-xtradb-operator/make.sh build
  ${REPO_ROOT}/hack/docker/percona-xtradb-operator/make.sh push
fi

if [ "$PROXYSQL_UPDATE" -eq 1 ]; then
  cowsay -f tux "Processing Proxysql images" || true
  for db in "${dbversions[@]}"; do
    ${REPO_ROOT}/hack/docker/proxysql/${db}/make.sh build
    ${REPO_ROOT}/hack/docker/proxysql/${db}/make.sh push
  done
fi
