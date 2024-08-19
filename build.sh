#!/usr/bin/env zsh

./tailwindcss -i pages/main.css -o static/main.css &&
  gzip < static/main.css > static/main.css.gz &&
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
