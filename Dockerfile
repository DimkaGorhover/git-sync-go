# syntax=docker/dockerfile:1.4-labs

ARG GOLANG_VERSION="1.19"
ARG BUILD_IMAGE="golang:${GOLANG_VERSION}-alpine"
ARG GOLANGCI_LINT_IMAGE="golangci/golangci-lint:latest"
ARG RELEASE_IMAGE="scratch"

# =============================================================================
FROM ${BUILD_IMAGE} as base

ARG APP_VERSION="docker"
LABEL \
    maintainer="Dmytro Horkhover <gd.mail.89@gmail.com>" \
    version="${APP_VERSION}"

SHELL ["/usr/bin/env", "sh", "-e", "-u" ,"-o", "pipefail", "-o", "errexit", "-o", "nounset", "-c"]

WORKDIR /src/git-sync

ARG GOARCH="amd64"
ARG GOOS="linux"
ENV GO111MODULE="on" \
    CGO_ENABLED="0"  \
    GOARCH="${GOARCH}" \
    GOOS="${GOOS}" \
    APP_VERSION="${APP_VERSION}"

COPY go.* .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# =============================================================================
FROM ${GOLANGCI_LINT_IMAGE} AS lint-base

# =============================================================================
FROM base AS lint

RUN --mount=target=. \
    --mount=from=lint-base,src=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/golangci-lint \
    golangci-lint run --timeout 10m0s ./...

# =============================================================================
FROM base AS test

RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go test -v -coverprofile=/cover.out ./... \
    && echo bingo > /lint.out

# =============================================================================
FROM base as build

RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build \
        -tags musl \
        -ldflags "-X 'main.AppVersion=${APP_VERSION}'" \
        -o /git-sync \
    && /git-sync --version

# wait until "test" stage is done
RUN --mount=from=test,src=/cover.out,target=/tmp/cover.out \
    --mount=from=test,src=/lint.out,target=/tmp/lint.out \
    cat /tmp/lint.out

# ============================================================================= 
FROM ${RELEASE_IMAGE} as release
LABEL maintainer="Dmytro Horkhover <gd.mail.89@gmail.com>"
COPY --from=build /git-sync /git-sync
ENTRYPOINT [ "/git-sync" ]
CMD [ "--version" ]
