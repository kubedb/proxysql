#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)
export KUBEDB_DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
export KUBEDB_NAMESPACE=${KUBEDB_NAMESPACE:-kube-system}
export MINIKUBE=0
export MINIKUBE_RUN=0
export SELF_HOSTED=1
export ARGS="" # Forward arguments to installer script

REPO_ROOT="$GOPATH/src/kubedb.dev/proxysql"
INSTALLER_ROOT="$GOPATH/src/github.com/kubedb/installer"

pushd $REPO_ROOT

# https://stackoverflow.com/a/677212/244009
if [[ ! -z "$(command -v onessl)" ]]; then
  export ONESSL=onessl
else
  # ref: https://stackoverflow.com/a/27776822/244009
  case "$(uname -s)" in
    Darwin)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.10.0/onessl-darwin-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    Linux)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.10.0/onessl-linux-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    CYGWIN* | MINGW32* | MSYS*)
      curl -fsSL -o onessl.exe https://github.com/kubepack/onessl/releases/download/0.10.0/onessl-windows-amd64.exe
      chmod +x onessl.exe
      export ONESSL=./onessl.exe
      ;;
    *)
      echo 'other OS'
      ;;
  esac
fi

source "$REPO_ROOT/hack/deploy/settings"

export KUBE_CA=$($ONESSL get kube-ca | $ONESSL base64)
export APPSCODE_ENV=${APPSCODE_ENV:-prod}
export KUBEDB_SCRIPT="curl -fsSL https://raw.githubusercontent.com/kubedb/installer/0.12.0/"

inside_git_repo() {
  git rev-parse --is-inside-work-tree >/dev/null 2>&1
  inside_git=$?
  if [ "$inside_git" -ne 0 ]; then
    echo "Not inside a git repository"
    exit 1
  fi
}

# Based on metadata() func in config.py
detect_tag() {
  inside_git_repo

  # http://stackoverflow.com/a/1404862/3476121
  git_tag=$(git describe --exact-match --abbrev=0 2>/dev/null || echo '')

  commit_hash=$(git rev-parse --verify HEAD)
  git_branch=$(git rev-parse --abbrev-ref HEAD)
  commit_timestamp=$(git show -s --format=%ct)

  if [ "$git_tag" != '' ]; then
    TAG=$git_tag
    TAG_STRATEGY='git_tag'
  elif [ "$git_branch" != 'master' ] && [ "$git_branch" != 'HEAD' ] && [[ "$git_branch" != release-* ]]; then
    TAG=$git_branch
    TAG_STRATEGY='git_branch'
  else
    hash_ver=$(git describe --tags --always --dirty)
    TAG="${hash_ver}"
    TAG_STRATEGY='commit_hash'
  fi

  echo "TAG = $TAG"
  echo "TAG_STRATEGY = $TAG_STRATEGY"
  echo "git_tag = $git_tag"
  echo "git_branch = $git_branch"
  echo "commit_hash = $commit_hash"
  echo "commit_timestamp = $commit_timestamp"

  # write TAG info to a file so that it can be loaded by a different command or script
  if [ $# -gt 0 ] && [ "$1" != '' ]; then
    cat >"$1" <<EOL
TAG=$TAG
TAG_STRATEGY=$TAG_STRATEGY
git_tag=$git_tag
git_branch=$git_branch
commit_hash=$commit_hash
commit_timestamp=$commit_timestamp
EOL
  fi
  export TAG
  export TAG_STRATEGY
  export git_tag
  export git_branch
  export commit_hash
  export commit_timestamp
}

show_help() {
  echo "setup.sh - setup kubedb operator"
  echo " "
  echo "setup.sh [options]"
  echo " "
  echo "options:"
  echo "-h, --help          show brief help"
  echo "    --selfhosted    deploy operator cluster."
  echo "    --minikube      setup configurations for minikube to run operator in localhost"
  echo "    --run           run operator in localhost and connect with minikube. only works with --minikube flag"
}

while test $# -gt 0; do
  case "$1" in
    -h | --help)
      show_help
      ARGS="$ARGS $1" # also show helps of "CLI repo" installer script
      shift
      ;;
    --docker-registry*)
      export KUBEDB_DOCKER_REGISTRY=$(echo $1 | sed -e 's/^[^=]*=//g')
      ARGS="$ARGS $1"
      shift
      ;;
    --minikube)
      export APPSCODE_ENV=dev
      export MINIKUBE=1
      export SELF_HOSTED=0
      shift
      ;;
    --run)
      export MINIKUBE_RUN=1
      shift
      ;;
    --selfhosted)
      export MINIKUBE=0
      export SELF_HOSTED=1
      shift
      ;;
    *)
      ARGS="$ARGS $1"
      shift
      ;;
  esac
done

# If APPSCODE_ENV==dev , use cli repo locally to run the installer script.
# Update "INSTALLER_BRANCH" in deploy/settings file to pull a particular CLI repo branch.
if [ "$APPSCODE_ENV" = "dev" ]; then
  detect_tag ''
  export KUBEDB_SCRIPT="cat $INSTALLER_ROOT/"
  export CUSTOM_OPERATOR_TAG=$TAG
  echo ""

  if [[ ! -d $INSTALLER_ROOT ]]; then
    echo ">>> Cloning cli repo"
    git clone -b $INSTALLER_BRANCH https://github.com/kubedb/cli.git "${INSTALLER_ROOT}"
    pushd $INSTALLER_ROOT
  else
    pushd $INSTALLER_ROOT
    detect_tag ''
    if [[ $git_branch != $INSTALLER_BRANCH ]]; then
      git fetch --all
      git checkout $INSTALLER_BRANCH
    fi
    git pull --ff-only origin $INSTALLER_BRANCH #Pull update from remote only if there will be no conflict.
    popd
  fi
fi

echo ""
env | sort | grep -e KUBEDB* -e APPSCODE*
echo ""

if [ "$SELF_HOSTED" -eq 1 ]; then
  echo "${KUBEDB_SCRIPT}deploy/kubedb.sh | bash -s -- --operator-name=proxysql-operator $ARGS"
  ${KUBEDB_SCRIPT}deploy/kubedb.sh | bash -s -- --operator-name=proxysql-operator ${ARGS}
fi

if [ "$MINIKUBE" -eq 1 ]; then
  cat $INSTALLER_ROOT/deploy/validating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
  cat $INSTALLER_ROOT/deploy/mutating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
  cat $REPO_ROOT/hack/dev/apiregistration.yaml | $ONESSL envsubst | kubectl apply -f -
#  cat $INSTALLER_ROOT/deploy/psp/proxysql.yaml | $ONESSL envsubst | kubectl apply -f -
  # Following line may give error if DBVersions CRD already not created
  cat $INSTALLER_ROOT/deploy/kubedb-catalog/proxysql.yaml | $ONESSL envsubst | kubectl apply -f - || true

  if [ "$MINIKUBE_RUN" -eq 1 ]; then
    go build -o ~/go/bin/proxysql-operator cmd/proxysql-operator/*.go
    proxysql-operator run --v=4 \
      --secure-port=8443 \
      --enable-status-subresource=true \
      --enable-mutating-webhook=true \
      --enable-validating-webhook=true \
      --kubeconfig="$HOME/.kube/config" \
      --authorization-kubeconfig="$HOME/.kube/config" \
      --authentication-kubeconfig="$HOME/.kube/config"
  fi
fi

if [ $(pwd) = "$INSTALLER_ROOT" ]; then
  popd
fi
popd
