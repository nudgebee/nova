FROM golang:1.26-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/nova *.go
RUN chmod +x /app/nova
RUN rm -rf /go

FROM alpine:3.24 AS release-stage

LABEL org.opencontainers.image.title="nova" \
      org.opencontainers.image.description="Find outdated or deprecated Helm charts running in your cluster (Nudgebee fork of FairwindsOps/nova)." \
      org.opencontainers.image.source="https://github.com/nudgebee/nova" \
      org.opencontainers.image.url="https://github.com/nudgebee/nova" \
      org.opencontainers.image.vendor="Nudgebee, Inc." \
      org.opencontainers.image.licenses="Apache-2.0" \
      io.nudgebee.upstream.source="https://github.com/FairwindsOps/nova"

WORKDIR /app
COPY --from=build-stage /app/nova /app/nova
RUN chmod +x /app/nova
