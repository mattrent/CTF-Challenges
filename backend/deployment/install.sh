#!/bin/bash

set -e

# Install KubeVirt
export RELEASE=$(curl https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt)
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${RELEASE}/kubevirt-operator.yaml
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${RELEASE}/kubevirt-cr.yaml
# kubectl -n kubevirt patch kubevirt kubevirt --type=merge --patch '{"spec":{"configuration":{"developerConfiguration":{"useEmulation":true}}}}'

# Install virtctl
if [[ $1 == "virtctl" ]]; then  
    export VERSION=$(curl https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt)
    wget https://github.com/kubevirt/kubevirt/releases/download/${VERSION}/virtctl-${VERSION}-linux-amd64
    sudo mv ./virtctl-${VERSION}-linux-amd64 /usr/local/bin/virtctl
    sudo chmod +x /usr/local/bin/virtctl
fi 

# Install Traefik
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install traefik traefik/traefik

# Install Helm Chart
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm dependency build ./helm
helm install -f ./helm/values.yaml deployer ./helm --namespace ctf --create-namespace
# helm upgrade deployer ./helm -f ./helm/values.yaml --namespace ctf
# helm uninstall deployer --namespace ctf
