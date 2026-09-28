# Echo service

A Go HTTP service that echoes any route and HTTP method. Responses have `headers`
and `params` as maps of string arrays, `body` as a raw string, and `path` as the
request path without its query string. Repeated values are preserved. `Host` is
included in headers. Standard HTTP semantics apply: HEAD responses have no body.

## Project layout

```text
cmd/echo-server/main.go   Application entry point and HTTP server setup
pkg/echo/                Echo handler, response model, and unit tests
infra/main.go            Pulumi configuration, provider, namespace, and outputs
infra/k8s/deployment.go  Kubernetes Deployment
infra/k8s/service.go     Kubernetes Service
infra/main_test.go       Pulumi resource wiring tests
```

## Prerequisites

Install Go 1.26.6 or newer, Docker with a running daemon, Kind, kubectl, and Pulumi.
The application needs Go 1.26.1; the Pulumi SDK needs Go 1.26.6. Older Go
installations with automatic toolchain downloads enabled will fetch it as needed.
The commands below run from the repository root unless otherwise indicated.
No cloud account or image registry is needed.

- [Docker installation](https://docs.docker.com/engine/install/)
- [Kind quick start](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [kubectl installation](https://kubernetes.io/docs/tasks/tools/)
- [Pulumi installation](https://www.pulumi.com/docs/install/)

## Run without Kubernetes

```sh
make test
make run
# In another terminal:
curl -H 'X-Example: hello' --data 'raw body' 'http://localhost:8080/anything?tag=a&tag=b'
```

Set `PORT=9090 make run` to override the default port. `make` lists available tasks.

## Create a cluster and build the image

Docker is the only supported container runtime. Run CI:

```sh
make ci
```

`make ci` tests and vets both Go modules, creates the configured Docker-backed
Kind cluster if missing, builds the Docker image, and loads it using
`kind load docker-image`. Existing clusters are reused.
You can also run `./scripts/ci.sh` directly.

The defaults are cluster `echo` and image `docker.io/library/web-server-eccho:local`.
Override the cluster or image when needed:

```sh
make ci KIND_CLUSTER_NAME=echo-docker IMAGE=docker.io/library/web-server-eccho:v2
```

The Deployment uses `imagePullPolicy: Never`, so the exact image tag must be
loaded into the selected cluster before deployment.

## Deploy with local Pulumi state

From the repository root, the Makefile handles local state and stack setup:

```sh
make pulumi-setup  # Create/select dev and configure image, cluster, and port
make preview
make deploy        # Run CI, load the image, and apply the deployment
make outputs
```

`make deploy` creates the Docker-backed Kind cluster if needed. For `make preview`,
create it first with `kind create cluster --name echo`. The default stack is `dev`
and the default cluster is `echo`. Override settings as needed, for example:

```sh
make deploy PULUMI_STACK=dev PORT=9090 IMAGE=docker.io/library/web-server-eccho:v2
```

The local-only passphrase defaults to `local-development`. For a different
stack passphrase, export `PULUMI_CONFIG_PASSPHRASE` before running Make. Setup
writes `IMAGE`, `KIND_CLUSTER_NAME`, and `PORT` into the selected stack config.
`make destroy` prompts before removing resources and retains the stack state
and Kind cluster.

Alternatively, run Pulumi directly:

```sh
cd infra
mkdir -p .pulumi
# Choose a local passphrase; reuse it for subsequent Pulumi commands.
export PULUMI_CONFIG_PASSPHRASE='local-development'
pulumi stack init dev
pulumi up
```

`Pulumi.yaml` selects a filesystem backend at `infra/.pulumi`; no Pulumi Cloud
login is required. Run all Pulumi commands from `infra/`. State and stack settings
are ignored by Git. Keep the state directory until the stack has been destroyed.
Use `pulumi stack select dev` when returning to an existing stack.

Optional settings, applied before `pulumi up`:

```sh
pulumi config set clusterName echo
pulumi config set image web-server-eccho:local
pulumi config set port 8080
```

The provider explicitly selects `kind-<clusterName>` from your kubeconfig. Pulumi
creates the fixed namespace `echo`, a Deployment with health probes, and a ClusterIP Service on
port 80. The configured application port is passed to the container and probes.

## Access the service

From `infra/`:

```sh
kubectl --context kind-echo -n "$(pulumi stack output namespace)" \
  port-forward "service/$(pulumi stack output service)" 8080:80
```

In another terminal:

```sh
curl -H 'X-Example: hello' --data 'raw body' 'http://localhost:8080/anything?tag=a&tag=b'
```

Example fields (curl also sends other headers):

```json
{"headers":{"Host":["localhost:8080"],"X-Example":["hello"]},"params":{"tag":["a","b"]},"body":"raw body","path":"/anything"}
```

For a custom cluster, replace `kind-echo` with its context name.

## Updating and CI

Prefer a new image tag on each update, so Kubernetes rolls out the new image:

```sh
# From the repository root:
IMAGE=docker.io/library/web-server-eccho:v2 ./scripts/ci.sh
cd infra
pulumi config set image docker.io/library/web-server-eccho:v2
pulumi up
```

GitHub Actions runs on pushes and pull requests. It creates Kind, executes the CI
script, deploys with local Pulumi state, checks a real echo response through the
Service, and destroys the infrastructure. CI uses the commit SHA as its image tag.

### Run the GitHub workflow locally with Docker

Install [act](https://nektosact.com/installation/index.html) and ensure
`docker run --rm hello-world` succeeds without sudo. If your user was just added
to the Docker group, log out and back in or run `newgrp docker` first.

The repository needs at least one commit because CI tags the image with the
commit SHA. On Linux, run:

```sh
make act-list
make act
# Optional overrides:
make act ACT_EVENT=pull_request
make act ACT_FLAGS=--pull=false
```

Host networking lets the runner reach Kind's API on the Docker host's loopback
address. The workflow creates and cleans up a separate Docker-backed cluster named
`echo-ci`, so your development cluster `echo` is unaffected. Avoid concurrent runs
using `echo-ci`. `ACT_IMAGE` overrides the runner image and `ACT_FLAGS` supplies
extra act arguments. Allow several GB of free
space for the runner, Kind image, and Go caches.

## Cleanup

Stop port-forwarding, then from `infra/`:

```sh
pulumi destroy
pulumi stack rm dev
kind delete cluster --name echo
```

## Publish the repository

Create an empty GitHub repository, then from the project root:

```sh
git add .
git commit -m "Implement echo service and local Kubernetes deployment"
git remote add origin https://github.com/YOUR-USER/web-server-eccho.git
git push -u origin HEAD
```

The resulting GitHub repository URL is the challenge submission link.
