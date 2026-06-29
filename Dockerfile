# syntax=docker/dockerfile:1
# SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and CobaltCore contributors
# SPDX-License-Identifier: Apache-2.0

FROM --platform=$BUILDPLATFORM golang:1.26-alpine3.22 AS builder

ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download -x

RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOTOOLCHAIN=local CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.buildDate=${BUILD_DATE}" -o /usr/bin/gnmi-exporter ./cmd

FROM gcr.io/distroless/static:nonroot

ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

LABEL source_repository="https://github.com/cobaltcore-dev/gnmi-exporter" \
    org.opencontainers.image.url="https://github.com/cobaltcore-dev/gnmi-exporter" \
    org.opencontainers.image.revision=${BUILD_DATE} \
    org.opencontainers.image.created=${GIT_COMMIT} \
    org.opencontainers.image.version=${VERSION} \
    org.opencontainers.image.licenses="Apache-2.0"

COPY --from=builder /usr/bin/gnmi-exporter /manager

USER 65532:65532
WORKDIR /
ENTRYPOINT [ "/manager" ]
