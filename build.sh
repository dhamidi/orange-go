#!/usr/bin/env zsh

# Compresses an asset and adds a checksum to the path, after the
# basename and before extensions.
# 
# prepare_asset "main.css" -> static/main.1234567.css.gz
prepare_asset() {
  local asset_path="$1"
  local asset_name=${asset_path##*/}
  asset_name=${asset_name%.*}
  local ext=${asset_path##*.}
  local checksum file_path
  gzip < static/"$asset_path" > static/"$asset_path".gz
  read checksum file_path < <(sha256sum static/"$asset_path".gz)
  mv -v static/"$asset_path".gz static/"$asset_name"."${checksum:0:7}"."$ext".gz
}

# Deletes all asset files with the same basename 
# 
# clean_asset main.css -> rm static/main.css static/main.1234567.css.gz
clean_asset() {
  find static -name "${1%.*}*" -exec rm -v {} \;
}

build_styles() {
  clean_asset main.css
  ./tailwindcss -i pages/main.css -o static/main.css
  prepare_asset main.css
}

# Write all asset files into a []string named IMMUTABLE_ASSETS in assets.go
collect_assets() {
  local outfile="assets.go"
  local asset asset_name ext
  printf "// AUTOGENERATED FILE, DO NOT EDIT\n\n" > "$outfile"
  printf "package main\n\n" >> "$outfile"
  printf "var IMMUTABLE_ASSETS = []string{\n" >> "$outfile"
  
  emit() {
    printf "found asset %s\n" "$1"
    printf "\t\"%q\",\n" "$1" >> $outfile
  }

  for asset in $@; do
    asset_name=${asset##*/}
    asset_name=${asset_name%.*}
    ext=${asset##*.}
    asset=$(find static -name "$asset_name.*.$ext.gz")
    if [[ -n "$asset" ]]; then 
      emit "${asset#static/}"
    fi
  done
  printf "}" >> $outfile

  gofmt -w assets.go
}

build_styles

collect_assets \
  htmx-sse.js \
  htmx.min.js \
  alpine-3.14.1.min.js \
  main.css


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
