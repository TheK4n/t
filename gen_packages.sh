#!/bin/sh

GOOS=linux GOARCH=amd64 go build -o ./t ./cmd/t
strip ./t
tar czf "t_v$(cat VERSION)_amd64.tar.gz" ./t