#!/usr/bin/env zsh
export GOEXPERIMENT=rangefunc

case "${1:-local}" in
"digitalocean")
  export CGO_ENABLED=1
  export GOARCH=amd64
  export GOOS=linux
  export CC=x86_64-linux-musl-gcc
  export CXX=x86_64-linux-musl-g++
  go build -ldflags "-linkmode external -extldflags -static" -o ./orange-linux .
  ;;
"local")
  go build .
  ;;
esac
