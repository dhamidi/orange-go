#!/usr/bin/env zsh
export GOEXPERIMENT=rangefunc
find . -name '*.go' | entr go test .
