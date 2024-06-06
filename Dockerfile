FROM --platform=linux/amd64 golang:1.22-alpine as build-stage


# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build  -o /app/nova *.go
RUN chmod +x /app/nova
RUN rm -rf /go

FROM alpine:3.19 AS release-stage
WORKDIR /app
COPY --from=build-stage /app/nova /app/nova
RUN chmod +x /app/nova