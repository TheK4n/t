#!/bin/sh

sed "s/%VERSION%/v$(cat VERSION)/g" ./.release/README.md.template > README.md
git tag -s -a "v$(cat VERSION)" -m "v$(cat VERSION)"