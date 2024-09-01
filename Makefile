SHELL :=/bin/bash

setup:
	@go mod tidy

build:
	@go build -v

install:build
	chmod +x S3Khoj
	sudo mv S3Khoj /usr/bin/
	sudo cp -r config/common-files.txt /usr/bin/