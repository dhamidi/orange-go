#!/usr/bin/env zsh
export GOEXPERIMENT=rangefunc
find . -name '*.go' | entr -rs 'go build && date && ./orange serve 127.0.0.1:8081'
