#!/bin/sh

log_execution() {
	name="$1"; shift

	echo "$(date -I'seconds'): $name - starting" | tee -a ./event-log.log >> /dev/stderr
	"$@"
	echo "$(date -I'seconds'): $name - ended with error code $?" | tee -a ./event-log.log >> /dev/stderr
}

set -a
[ -f .env ] && . ./.env
set +a

while true; do
	mkdir -p state

	log_execution 'fetcher' \
		./fetcher

	rm -rfv to-upload

	log_execution 'apifier' \
		./apifier -output-dir to-upload

	log_execution 'uploader' \
		./uploader \
			-directory to-upload \
			-bucket-name "${R2_PUBLIC_BUCKET_NAME:?}"

	rm -rfv to-upload

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
done

