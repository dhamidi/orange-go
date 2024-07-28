#!/usr/bin/env zsh
export GOEXPERIMENT=rangefunc
find . -name '*.go' | entr -crs 'go build && date && ./orange serve 127.0.0.1:8081'
