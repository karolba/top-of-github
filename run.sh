#!/bin/sh
while true; do
	./fetcher || { echo "[!!] Fetcher failed, return code: $?"; }
	rm -rfv to-upload
	./apifier -output-dir to-upload
	./uploader -directory to-upload
done

