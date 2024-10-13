
# CTF

## Prerequisites

- Kubernetes cluster with KubeVirt and Traefik installed.
- Helm client installed.

## Backend

Update the configuration `backend/deployment/helm/values.yaml` and install the Helm chart.

A default login will be created with a random password. To get the password use the command: `kubectl get secrets/deployer --template={{.data.password}} | base64 -d`.

See `backend/examples/requests.http` for examples of API usage.

### Frontend

Challenges can be deployed from CTFd with `frontend/container_challenges`.

The plugin requires setting the environment variables `JWTSECRET` and `BACKENDURL`, where `JWTSECRET` will be available in the secret mentioned above.

