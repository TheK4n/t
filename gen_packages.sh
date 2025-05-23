#!/bin/sh


build() {
    arch="${1}"
    go generate ./...
    go fmt ./...
    GOOS=linux GOARCH="${arch}" go build -ldflags "-w -s" -o ./t ./cmd/t
}

build_android_arm64() {
    go generate ./...
    go fmt ./...
    CGO_ENABLED=1 \
    GOOS="android" \
    GOARCH="arm64" \
    CC="aarch64-linux-android21-clang" \
    go build -ldflags "-w -s" -o ./t ./cmd/t
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

build_and_pack_android() {
    build_android_arm64
    tar czf "t_v$(cat VERSION)-android_arm64.tar.gz" ./t
}

build_and_pack "arm64"
build_and_pack "amd64"

build_and_pack_android
