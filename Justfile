static-linux-apifier:
	mkdir -p linux-build
	cd apifier && CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v --ldflags '-linkmode external -extldflags "-static"' -tags 'osusergo netgo static_build' -o ../linux-build/apifier

static-linux-fetcher:
	mkdir -p linux-build
	cd fetcher && CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v --ldflags '-linkmode external -extldflags "-static"' -tags 'osusergo netgo static_build' -o ../linux-build/fetcher

static-linux-uploader:
	mkdir -p linux-build
	cd uploader && GOOS=linux GOARCH=amd64 go build -v -o gtkf_linux -tags 'osusergo netgo static_build' -o ../linux-build/uploader

static-linux: static-linux-apifier static-linux-fetcher static-linux-uploader
