
# CTF

## Prerequisites

- Kubernetes cluster with KubeVirt installed.
- Helm client installed.

## Frontend

Challenges can be deployed from CTFd with the plugin in `frontend/container_challenges`.

The plugin requires the environment variable `BACKENDURL` to call the Backend API.

The optional environment variable `API_TOKEN` will create an initial API token in CTFd. The API token should have same value in both the frontend and backend configuration if deploying both applications in one operation.

## Backend

Update the configuration `backend/deployment/helm/values.yaml` and install the Helm chart.

A default login will be created with a random password. To get the password use the command: `kubectl get secrets/deployer --template={{.data.password}} | base64 -d`.

See `backend/examples/requests.http` for examples of API usage.

## Challenge examples

Challenge examples are found in `backend/examples/`.

The API allows adding, updating, starting, and stopping challenges. After adding a challenge, it can be deployed to CTFd using the publish API endpoint. With the CTFd plugin installed, players can start and stop published challenges from CTFd. See the scripts in the deployment directory of each challenge for examples of how to deploy and publish challenges.

## Development

API Documentation at: `/swagger/index.html`

Generate Swagger documentation: `swag init -g ./cmd/server/main.go -o ./docs`
