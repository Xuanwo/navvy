SHELL := /bin/bash

.PHONY: all check formatvet lint test

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  check      to format, vet and lint "
	@echo "  build      to create bin directory and build nomad-driver-oci"
	@echo "  install    to install nomad-driver-oci to /usr/local/bin/nomad-driver-oci"
	@echo "  release    to release nomad-driver-oci"
	@echo "  test       to run test"

check: format vet lint

format:
	@echo "go fmt code"
	@go fmt ./...
	@echo "ok"

vet:
	@echo "go vet code"
	@go vet -all ./...
	@echo "ok"

lint:
	@echo "golint code"
	@golint ./...
	@echo "ok"

test:
	@echo "run test"
	@go test -v ./...
	@echo "ok"
