#!/data/data/com.termux/files/usr/bin/bash

export BOOTSTRAP="${BASH_SOURCE[0]}"
export SRCDIR=".."
export WRAPPER="../blueprint.bash"

../bootstrap.bash "$@"
