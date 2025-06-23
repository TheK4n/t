#!/bin/sh

version="$(git describe --tags --abbrev=0)"
go build -ldflags "-w -s -X 'main.version=${version}'" -o bin/ ./cmd/t
