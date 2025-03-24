#!/bin/bash
set -eu -o pipefail

cd "$(dirname "$0")/.."

if ! [[ $# == 1 ]]; then
    echo "Usage: $0 [docker remote context name]" >&2
    exit 1
fi
remote_context="$1"

with-log() (set -x; "$@")
local-docker() { with-log docker "$@"; }
remote-docker() { with-log docker --context "$remote_context" "$@"; }

local-docker build -t top-of-github --platform linux/amd64 .
local-docker save top-of-github | remote-docker load
remote-docker compose up -d
