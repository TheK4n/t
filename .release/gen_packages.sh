#!/bin/sh

GOOS=linux GOARCH=amd64 go build -o ./t ./cmd/t
strip ./t
tar czf "t_v$(cat VERSION)_linux_amd64.tar.gz" ./t