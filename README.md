
# CTF Challenge Deployer

## Prerequesites

- Kubernetes cluster with KubeVirt and Traefik installed.
- Helm client installed.

## Installation

Update the configuration `deployment/helm/values.yaml` and install the Helm chart.

A default login will be created with a random password. To get the password use the command: `kubectl get secrets/deployer --template={{.data.password}} | base64 -d`.

See `examples/requests.http` for examples of API usage.

### CTFd Plugin

Challenges can be deployed from CTFd with `ctfd-plugin/container_challenges`.

The plugin requires setting the environment variables `jwtsecret` and `BACKENDURL`, where `jwtsecret` will be available in the secret mentioned above.

