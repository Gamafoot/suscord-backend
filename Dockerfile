# syntax=docker/dockerfile:1.7

FROM golang:1.25-bookworm AS builder

WORKDIR /build

ARG TARGETOS
ARG TARGETARCH

COPY src/go.mod src/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY src .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -o /out/api ./cmd/api \
    && \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -o /out/migrate ./cmd/migrate


FROM debian:bookworm-slim

WORKDIR /app

RUN groupadd --system nonroot \
    && useradd --system --gid nonroot --home-dir /app --shell /usr/sbin/nologin nonroot

COPY --chown=nonroot:nonroot entrypoint.sh /app/entrypoint.sh
COPY --chown=nonroot:nonroot config /app/config
COPY --chown=nonroot:nonroot src/assets /app/assets
COPY --chown=nonroot:nonroot --from=builder /out/api /app/api
COPY --chown=nonroot:nonroot --from=builder /out/migrate /app/migrate

RUN chmod +x /app/entrypoint.sh /app/api /app/migrate

USER nonroot

ENV CONFIG_PATH=/app/config/config.yaml

ENTRYPOINT ["/bin/sh", "/app/entrypoint.sh"]
