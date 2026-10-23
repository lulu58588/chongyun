#!/bin/sh
# Sign a binary or an app bundle for the Mac, ad-hoc.
#
# An unsigned binary the Mac has not run before goes through the Gatekeeper's
# check on its first run, and that check asks Apple's servers: on a network
# that cannot reach them, the binary hangs (the process sits in an
# uninterruptible wait, printing nothing) instead of starting. An ad-hoc
# signature settles it locally, so the binary runs at once.
#
# A Developer ID signature is better still (that is what a release carries);
# this is what a local build gets.
set -e

case "$(uname -s)" in
Darwin) ;;
*)
	exit 0
	;;
esac

target="$1"
if [ -z "$target" ] || [ ! -e "$target" ]; then
	echo "  (nothing to sign: $target)" >&2
	exit 0
fi

if codesign --force --sign - "$target" >/dev/null 2>&1; then
	echo "  signed (ad-hoc): $target"
else
	echo "  (could not sign $target; it still runs, but the Mac may check it on first run)" >&2
fi
