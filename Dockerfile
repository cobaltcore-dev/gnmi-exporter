# syntax=docker/dockerfile:1
# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and CobaltCore contributors
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.25-alpine3.21 AS builder

RUN apk add --no-cache --no-progress git make openssh

ARG BININFO_BUILD_DATE
ARG BININFO_COMMIT_HASH
ARG BININFO_VERSION

ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

ARG GITHUB_TOKEN
RUN if [ -n "$GITHUB_TOKEN" ]; then \
        git config --global url."https://x-oauth-basic:${GITHUB_TOKEN}@github.com/ironcore-dev".insteadOf "https://github.com/ironcore-dev"; \
    else \
        git config --global url."git@github.com:ironcore-dev".insteadOf "https://github.com/ironcore-dev" \
        && mkdir -p /root/.ssh \
        && ssh-keyscan github.com >> /root/.ssh/known_hosts; \
    fi

ENV GOPRIVATE=github.com/ironcore-dev/*
RUN --mount=type=ssh \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download -x

RUN --mount=type=bind,target=.,readwrite \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOTOOLCHAIN=local make PREFIX=/pkg install

FROM gcr.io/distroless/static:nonroot

ARG BININFO_BUILD_DATE
ARG BININFO_COMMIT_HASH
ARG BININFO_VERSION

LABEL source_repository="https://github.com/cobaltcore-dev/monitoring-operator" \
    org.opencontainers.image.url="https://github.com/cobaltcore-dev/monitoring-operator" \
    org.opencontainers.image.created=${BININFO_BUILD_DATE} \
    org.opencontainers.image.revision=${BININFO_COMMIT_HASH} \
    org.opencontainers.image.version=${BININFO_VERSION}

COPY --from=builder /pkg/ /usr/

USER 65532:65532
WORKDIR /
ENTRYPOINT [ "/usr/bin/monitoring-operator" ]
