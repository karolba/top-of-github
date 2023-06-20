FROM alpine:3.18.2

ADD run.sh linux-build/apifier linux-build/fetcher linux-build/uploader /opt/top-of-github/

WORKDIR /opt/top-of-github
ENTRYPOINT ./run.sh
