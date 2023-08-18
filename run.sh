#!/bin/sh

log_execution() {
	name="$1"; shift

	echo "$(date -I'seconds'): $name - starting" | tee -a state/event-log.log >> /dev/stderr
	"$@"
	exit_code=$?
	echo "$(date -I'seconds'): $name - ended with error code $exit_code" | tee -a state/event-log.log >> /dev/stderr

	if [ "$exit_code" -ne 0 ]; then
		exit "$exit_code"
	fi
}

set -a
[ -f .env ] && . ./.env
set +a

while true; do
	mkdir -p state

	log_execution 'fetcher' \
		./fetcher

	rm -rf to-upload

	log_execution 'apifier' \
		./apifier -output-dir to-upload

	log_execution 'uploader' \
		./uploader \
			-directory to-upload \
			-bucket-name "${R2_PUBLIC_BUCKET_NAME:?}"

	rm -rf to-upload

	log_execution 'sqlite vacuum' \
		sqlite3 state/repos.db \
			"pragma journal_mode=WAL; pragma temp_store_directory='.'; vacuum"

	rm -rfv to-upload-backup
	mkdir -p to-upload-backup

	log_execution 'compressing state.db for a backup upload' \
		gzip < state/repos.db > to-upload-backup/repos.db.gz

	log_execution 'uploading compressed database backup' \
		./uploader \
			-bucket-name "${R2_DB_BACKUP_BUCKET_NAME:?}" \
			-directory to-upload-backup \
			-content-type application/x-sqlite3

	rm -rfv to-upload-backup

	log_execution 'purging the metadata file from cache' \
		curl -X DELETE "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_PURGE_CACHE_ZONE:?}/purge_cache" \
		-H "Authorization: bearer ${CLOUDFLARE_PURGE_CACHE_TOKEN:?}" \
		-H "Content-Type: application/json" \
		--data "{\"files\":[\"https://${CLOUDFLARE_PURGE_CACHE_DOMAIN:?}/metadata\"]}"

done

