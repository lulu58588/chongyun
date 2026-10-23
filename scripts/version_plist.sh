#!/bin/sh
# The version macOS shows in its places: three numbers and nothing else.
# CFBundleShortVersionString accepts nothing but that, so the build's own
# version (0.2.0-dev.9.dirty and the like) goes in CFBundleVersion instead.
set -e

cd "$(dirname "$0")/.."

sh scripts/version.sh | sed -E 's/^v?([0-9]+(\.[0-9]+){0,2}).*$/\1/'
