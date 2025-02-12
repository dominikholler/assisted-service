#!/usr/bin/env bash

set -o nounset
set -o pipefail
set -o errexit
set -o xtrace

curl -L "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" -o /tmp/kubectl && \
    install -o root -g root -m 0755 /tmp/kubectl /usr/local/bin/kubectl && \
    rm -f /tmp/kubectl
yum install -y docker libvirt-clients awscli python3-pip postgresql genisoimage && \
    yum clean all
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | \
    bash -s -- 3.8.8 && mv kustomize /usr/bin/
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.36.0
curl -L https://raw.githack.com/stoplightio/spectral/master/scripts/install.sh | sh

ARCH=$(case $(arch) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(arch) ;; esac)
OS=$(uname | awk '{print tolower($0)}')
OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.6.2
curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}
chmod +x operator-sdk_${OS}_${ARCH}
install operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk

go get -u github.com/onsi/ginkgo/ginkgo@v1.16.1 \
    golang.org/x/tools/cmd/goimports@v0.1.0 \
    github.com/golang/mock/mockgen@v1.4.3 \
    github.com/vektra/mockery/.../@v1.1.2 \
    gotest.tools/gotestsum@v1.6.3 \
    github.com/axw/gocov/gocov \
    sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.0 \
    github.com/AlekSi/gocov-xml@v0.0.0-20190121064608-3a14fb1c4737

python3 -m pip install --upgrade pip
python3 -m pip install -r ./dev-requirements.txt
