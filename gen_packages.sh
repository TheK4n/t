#!/bin/sh


build() {
    arch="${1}"
    go generate ./...
    go fmt ./...
    GOOS=linux GOARCH="${arch}" go build -ldflags "-w -s" -o ./t ./cmd/t
}

build_sqlite() {
    arch="${1}"
    go generate ./...
    go fmt ./...
    GOOS=linux GOARCH="${arch}" go build --tags tsqlite -ldflags "-w -s" -o ./t ./cmd/t
}

build_sqlite_android_arm64() {
    go generate ./...
    go fmt ./...
    CGO_ENABLED=1 \
    GOOS="android" \
    GOARCH="arm64" \
    CC="aarch64-linux-android21-clang" \
    go build --tags tsqlite -ldflags "-w -s" -o ./t ./cmd/t
}

pack() {
    arch="${1}"
    tar czf "t_v$(cat VERSION)_linux_${arch}.tar.gz" ./t
}

pack_sqlite() {
    arch="${1}"
    tar czf "t_v$(cat VERSION)-sqlite_linux_${arch}.tar.gz" ./t
}

build_and_pack() {
    arch="${1}"
    build "${arch}"
    pack "${arch}"
}

build_and_pack_sqlite() {
    arch="${1}"
    build_sqlite "${arch}"
    pack_sqlite "${arch}"
}

build_and_pack_sqlite_android() {
    build_sqlite_android_arm64
    tar czf "t_v$(cat VERSION)-sqlite_android_arm64.tar.gz" ./t
}

build_and_pack_sqlite "amd64"

build_and_pack "arm64"
build_and_pack "amd64"

build_and_pack_sqlite_android
