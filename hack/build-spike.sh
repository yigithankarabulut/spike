#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

rm keeper
rm nexus
rm spike

# `boringcrypto` is required for FIPS compliance.

CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper ./app/keeper/cmd/main.go
CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o nexus ./app/nexus/cmd/main.go
CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike ./app/spike/cmd/main.go

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper-x86 ./app/keeper/cmd/main.go
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o nexus-x68 ./app/nexus/cmd/main.go
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike-x86 ./app/spike/cmd/main.go

GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper-darwin ./app/keeper/cmd/main.go
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o nexus-darwin ./app/nexus/cmd/main.go
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike-darwin ./app/spike/cmd/main.go

GOOS=linux GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper-arm ./app/keeper/cmd/main.go
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o nexus-arm ./app/nexus/cmd/main.go
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike-arm ./app/spike/cmd/main.go