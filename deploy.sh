#!/usr/bin/env zsh

set -eux

deploy_user=$(git config deploy.user)
deploy_host=$(git config deploy.host)
deploy_path=$(git config deploy.path)
case "${1:-push}" in
  push)
    ./build.sh digitalocean
    scp ./orange-linux "$deploy_user"@"$deploy_host":"$deploy_path.next"
    ;;
  shell)
    exec ssh "$deploy_user"@"$deploy_host"
    ;;
esac
