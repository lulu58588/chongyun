#!/bin/sh
# The version of this build.
#
# Starling has its own line, from 0.2.0 on: the tags this repository carries
# from upstream (v0.1.x) are not ours and must not show up as our version.
# Before the first v0.2.x tag exists, the build says 0.2.0-dev.<commits>, so
# a build that is not a release still says which line it is from.
set -e

cd "$(dirname "$0")/.."

describe=$(git describe --tags --match 'v0.2.*' --always --dirty 2>/dev/null || true)

case "$describe" in
v0.2.*)
	# A tag of our own line: use it as it is.
	echo "$describe" | tr -cd 'A-Za-z0-9._+-'
	;;
*)
	# No tag of ours yet: say the line, the commits since, and whether the
	# tree is dirty.
	commits=$(git rev-list --count HEAD 2>/dev/null || echo 0)
	if git diff --quiet 2>/dev/null && git diff --cached --quiet 2>/dev/null; then
		echo "0.2.0-dev.${commits}"
	else
		echo "0.2.0-dev.${commits}.dirty"
	fi
	;;
esac
