#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
export PATH="$(go env GOPATH)/bin:${PATH}"
rm -rf ../gen/go
buf generate --template buf.gen.getway.yaml --path userService.proto --path train_service_api
mkdir -p ../gen/go/userServicepb
mv ../gen/go/userService*.go ../gen/go/userServicepb/
