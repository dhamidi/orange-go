#!/usr/bin/env zsh

if ! [[ -f .env ]]; then
  touch .env
fi
find . -name '*.go' | entr -crs '
  ./build.sh &&
  date && 
  set -a && 
  . ./.env && 
  ./orange serve 127.0.0.1:8081
'
