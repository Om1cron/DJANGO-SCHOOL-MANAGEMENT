#!/bin/bash

# builds a macos package (.app) and places it inside a compressed .dmg

export CGO_CFLAGS="-mmacosx-version-min=10.14"
export CGO_LDFLAGS="-mmacosx-version-min=10.14"

mkdir -p "package/Cryptonym"
mkdir -p package/old
mv package/*.dmg package/old/

go build -ldflags "-s -w" -o cmd/cryptonym-wallet/c