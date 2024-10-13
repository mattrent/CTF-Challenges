#!/bin/bash

# Install KubeVirt
export RELEASE=$(curl https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt)
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${RELEASE}/kubevirt-operator.yaml
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${RELEASE}/kubevirt-cr.yaml

# Install Traefik
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install traefik traefik/traefik

# Install Helm Chart
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm dependency build ./helm
helm install -f ./helm/values.yaml deployer ./helm
