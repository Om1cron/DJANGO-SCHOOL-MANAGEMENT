#!/bin/bash

# builds a macos package (.app) and places it inside a compressed .dmg

export CGO_CFLAGS="-mmacosx-version-min=10.14"
export CGO_LDFLAGS="-mmacosx-version-