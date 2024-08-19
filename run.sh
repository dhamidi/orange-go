#!/usr/bin/env zsh

if ! [[ -f .env ]]; then
  touch .env
fi
find . -name '*.go' | entr -crs '
  ./tailwindcss -i pages/main.css -o static/main.css &&
  gzip < static/main.css > static/main.css.gz &&
  go build && 
  date && 
  set -a && 
  . ./.env && 
  ./orange serve 127.0.0.1:8081
'
