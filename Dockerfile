# Build
FROM golang:1.22-alpine AS build-env
RUN apk add build-base
WORKDIR /app
COPY . /app
RUN go mod tidy
RUN go build -o S3Khoj main.go

# Release
FROM alpine:3.18.6
RUN apk upgrade --no-cache \
    && apk add --no-cache bind-tools 
COPY --from=build-env /app/S3Khoj /usr/local/bin/S3Khoj

ENTRYPOINT [ "S3Khoj" ]