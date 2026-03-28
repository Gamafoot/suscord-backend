FROM golang:1.25-alpine AS builder

WORKDIR /build

RUN apk add --no-cache build-base

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src .

RUN CGO_ENABLED=1 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=1 go build -o /out/migrate ./cmd/migrate


FROM alpine:3.19

WORKDIR /app

RUN addgroup -S nonroot
RUN adduser -S nonroot -G nonroot
RUN chown -R nonroot:nonroot /app

COPY entrypoint.sh entrypoint.sh
COPY config config
COPY --from=builder /out/api api
COPY --from=builder /out/migrate migrate
COPY src/assets assets

RUN mkdir database
RUN chmod -R 777 database
RUN chmod +x entrypoint.sh

USER nonroot

ENV CONFIG_PATH=/app/config/config.yaml

ENTRYPOINT ["/bin/sh", "entrypoint.sh"]
