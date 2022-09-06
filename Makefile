SHELL  := /usr/bin/env bash -e -u -o pipefail -o errexit -o nounset

export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

.SILENT:
build:
	docker build \
		--force-rm \
    	--rm \
    	--target release \
		-t dier/git-sync-go:latest \
		$(shell pwd)
