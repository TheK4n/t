#!/bin/sh


build() {
    arch="${1}"
    go generate ./...
    GOOS=linux GOARCH="${arch}" go build -ldflags "-w -s" -o ./t ./cmd/t
}

pack() {
    arch="${1}"
    tar czf "t_v$(cat VERSION)_linux_${arch}.tar.gz" ./t
}

build_and_pack() {
    arch="${1}"
    build "${arch}"
    pack "${arch}"
}

build_and_pack "amd64"
build_and_pack "arm64"