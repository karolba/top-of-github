#!/bin/sh
set -a
source .env
set +a
while true; do
	./fetcher || { echo "[!!] Fetcher failed, return code: $?"; }
	rm -rfv to-upload
	./apifier -output-dir to-upload
	./uploader -directory to-upload
done

