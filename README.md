
# CTF

## Prerequisites

- Kubernetes cluster with KubeVirt and Traefik installed.
- Helm client installed.

## Backend

Update the configuration `backend/deployment/helm/values.yaml` and install the Helm chart.

A default login will be created with a random password. To get the password use the command: `kubectl get secrets/deployer --template={{.data.password}} | base64 -d`.

See `backend/examples/requests.http` for examples of API usage.

## Frontend

Challenges can be deployed from CTFd with the plugin in `frontend/container_challenges`.

The plugin requires the environment variable `BACKENDURL`.

## Challenge examples

Challenge examples can be found in `backend/examples/`.

The API allows adding, updating, starting, and stopping challenges. After adding a challenge, it can be deployed to CTFd using the publish API endpoint. With the CTFd plugin installed, players can start and stop published challenges from CTFd. See the scripts in the deployment directory of each challenge for examples of how to deploy and publish challenges.
