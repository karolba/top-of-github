#!/bin/sh

log() {
        echo "`date -I'seconds'`: $*" | tee -a ./event-log.log
}

set -a
source .env
set +a

mkdir -p state

while true; do
        log "fetcher: starting"
        ./fetcher
        log "fetcher: ended with code $?"

        rm -rfv to-upload

        log "apifier: starting"
        ./apifier -output-dir to-upload
        log "apifier: ended with code $?"

        log "uploader: starting"
        ./uploader -directory to-upload
        log "uploader: ended with code $?"
done

