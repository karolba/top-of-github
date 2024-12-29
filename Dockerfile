ARG ALPINE_VERSION=3.18
ARG GO_VERSION=1.21

FROM public.ecr.aws/docker/library/golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder-base
RUN --mount=type=cache,target=/var/cache/apk,sharing=private \
    apk add --update gcc g++ musl-dev
WORKDIR /build

FROM builder-base AS apifier-builder
COPY apifier/go.mod apifier/go.sum .
RUN --mount=type=cache,target=/go/mod/pkg/cache,sharing=shared \
    go mod download
COPY apifier/*.go .
RUN --mount=type=cache,target=/root/.cache/go-build,sharing=shared \
    CGO_ENABLED=1 go build -v -o apifier --ldflags '-linkmode external -extldflags "-static"'

FROM builder-base AS uploader-builder
COPY uploader/go.mod uploader/go.sum .
RUN --mount=type=cache,target=/go/mod/pkg/cache,sharing=shared \
    go mod download
COPY uploader/*.go .
RUN --mount=type=cache,target=/root/.cache/go-build,sharing=shared \
    CGO_ENABLED=1 go build -v -o uploader --ldflags '-linkmode external -extldflags "-static"'

FROM builder-base AS fetcher-builder
COPY fetcher/go.mod fetcher/go.sum .
RUN --mount=type=cache,target=/go/mod/pkg/cache,sharing=shared \
    go mod download
COPY fetcher/*.go .
RUN --mount=type=cache,target=/root/.cache/go-build,sharing=shared \
    CGO_ENABLED=1 go build -v -o fetcher --ldflags '-linkmode external -extldflags "-static"'

FROM public.ecr.aws/docker/library/alpine:${ALPINE_VERSION} AS runner
RUN --mount=type=cache,target=/var/cache/apk,sharing=private \
    apk add --update sqlite curl
WORKDIR /top-of-github
COPY --link --from=apifier-builder /build/apifier .
COPY --link --from=fetcher-builder /build/fetcher .
COPY --link --from=uploader-builder /build/uploader .
COPY run.sh .
VOLUME ["/top-of-github/state"]
CMD ["./run.sh"]
