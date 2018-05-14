#!/data/data/com.termux/files/usr/bin/bash

set -e

[ -z "$BUILDDIR" ] && BUILDDIR=`dirname "${BASH_SOURCE[0]}"`

[ -z "$NINJA" ] && NINJA=ninja


if [ ! -f "${BUILDDIR}/.blueprint.bootstrap" ]; then
    echo "Please run bootstrap.bash (.blueprint.bootstrap missing)" >&2
    exit 1
fi

source "${BUILDDIR}/.blueprint.bootstrap"

if [ -z "$BLUEPRINTDIR" ]; then
    echo "Please run bootstrap.bash (.blueprint.bootstrap outdated)" >&2
    exit 1
fi

source "${BLUEPRINTDIR}/blueprint_impl.bash"
