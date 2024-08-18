#!/usr/bin/env zsh

set -eux

deploy_user=$(git config deploy.user)
deploy_host=$(git config deploy.host)
deploy_path=$(git config deploy.path)
version=$(date +%Y%m%d%H%M%S)
case "${1:-push}" in
  push)
    ./build.sh digitalocean
    scp ./orange-linux "$deploy_user"@"$deploy_host":"$deploy_path.$version"
    ;;
  shell)
    exec ssh "$deploy_user"@"$deploy_host"
    ;;
esac
