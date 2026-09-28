#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

export IMAGE="${IMAGE:-docker.io/library/web-server-eccho:local}"
export KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-echo}"
# Always load into Docker-backed Kind, including with stale shell settings.
export KIND_EXPERIMENTAL_PROVIDER=docker

for tool in go docker kind; do
  command -v "$tool" >/dev/null || {
    echo "$tool is required" >&2
    exit 1
  }
done

go test -race ./...
go vet ./...
(cd infra && go test ./... && go vet ./...)

clusters=$(kind get clusters)
if ! printf '%s\n' "$clusters" | grep -Fxq -- "$KIND_CLUSTER_NAME"; then
  kind create cluster --name "$KIND_CLUSTER_NAME" --wait 120s
fi

docker build -t "$IMAGE" .
kind load docker-image "$IMAGE" --name "$KIND_CLUSTER_NAME"
printf 'Loaded %s into Kind cluster %s\n' "$IMAGE" "$KIND_CLUSTER_NAME"
