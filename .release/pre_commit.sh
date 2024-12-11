#!/bin/sh

go generate ./...
sed "s/%VERSION%/v$(cat VERSION)/g" ./.release/README.md.template > README.md