SHELL :=/bin/bash

setup:
	@go mod tidy

build:
	@go build -v
