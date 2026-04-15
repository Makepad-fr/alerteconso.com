# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.24.3

FROM golang:${GO_VERSION} AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/alerteconso ./main.go

FROM alpine:3.21 AS runtime

ARG BUILD_DATE=1970-01-01T00:00:00Z
ARG VCS_REF=local
ARG VERSION=local

RUN apk add --no-cache ca-certificates

WORKDIR /app

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
	org.opencontainers.image.revision="${VCS_REF}" \
	org.opencontainers.image.version="${VERSION}"

COPY --from=builder /out/alerteconso ./alerteconso
COPY templates ./templates

EXPOSE 8080

CMD ["./alerteconso"]
