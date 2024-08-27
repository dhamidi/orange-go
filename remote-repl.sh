#!/usr/bin/env zsh

rlwrap=rlwrap
if ! which rlwrap 2>&1 >/dev/null; then
  printf "* rlwrap is not installed. You shall suffer.\n" >&2
  rlwrap=''
fi

ssh -f -L 12345:localhost:8088 $(git config deploy.user)@$(git config deploy.host) sleep 10
$rlwrap nc localhost 12345
