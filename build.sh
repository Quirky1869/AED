#!/usr/bin/env bash
set -euo pipefail

# build
mkdir -p bin

go build -o bin/aed ./aed.go
