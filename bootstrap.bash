#!/data/data/com.termux/files/usr/bin/bash
set -e

EXTRA_ARGS=""

if [ -z "$BOOTSTRAP" ]; then
    BOOTSTRAP="${BASH_SOURCE[0]}"
    [ -z "$WRAPPER" ] && WRAPPER="`dirname "${BOOTSTRAP}"`/blueprint.bash"
fi

[ -z "$SRCDIR" ] && SRCDIR=`dirname "${BOOTSTRAP}"`
[ -z "$BLUEPRINTDIR" ] && BLUEPRINTDIR="${SRCDIR}"
[ -z "$BUILDDIR" ] && BUILDDIR=.
[ -z "$NINJA_BUILDDIR" ] && NINJA_BUILDDIR="${BUILDDIR}"
[ -z "$TOPNAME" ] && TOPNAME="Blueprints"
[ -z "$GOROOT" ] && GOROOT=`go env GOROOT`

usage() {
    echo "Usage of ${BOOTSTRAP}:"
    echo "  -h: print a help message and exit"
    echo "  -b <builddir>: set the build directory"
    echo "  -t: run tests"
}

while getopts ":b:ht" opt; do
    case $opt in
        b) BUILDDIR="$OPTARG";;
        t) RUN_TESTS=false;;
        h)
            usage
            exit 1
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            usage
            exit 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            exit 1
            ;;
    esac
done

[ ! -z "$RUN_TESTS" ] && EXTRA_ARGS="${EXTRA_ARGS}"

if [ -z "${BLUEPRINT_LIST_FILE}" ]; then
  BLUEPRINT_LIST_FILE="${BUILDDIR}/.bootstrap/bplist"
fi
EXTRA_ARGS="${EXTRA_ARGS} -l ${BLUEPRINT_LIST_FILE}"

mkdir -p $BUILDDIR/.minibootstrap

echo "bootstrapBuildDir = $BUILDDIR" > $BUILDDIR/.minibootstrap/build.ninja
echo "topFile = $SRCDIR/$TOPNAME" >> $BUILDDIR/.minibootstrap/build.ninja
echo "extraArgs = $EXTRA_ARGS" >> $BUILDDIR/.minibootstrap/build.ninja
echo "builddir = $NINJA_BUILDDIR" >> $BUILDDIR/.minibootstrap/build.ninja
echo "include $BLUEPRINTDIR/bootstrap/build.ninja" >> $BUILDDIR/.minibootstrap/build.ninja

echo "BLUEPRINT_BOOTSTRAP_VERSION=2" > $BUILDDIR/.blueprint.bootstrap
echo "SRCDIR=\"${SRCDIR}\"" >> $BUILDDIR/.blueprint.bootstrap
echo "BLUEPRINTDIR=\"${BLUEPRINTDIR}\"" >> $BUILDDIR/.blueprint.bootstrap
echo "NINJA_BUILDDIR=\"${NINJA_BUILDDIR}\"" >> $BUILDDIR/.blueprint.bootstrap
echo "GOROOT=\"${GOROOT}\"" >> $BUILDDIR/.blueprint.bootstrap
echo "TOPNAME=\"${TOPNAME}\"" >> $BUILDDIR/.blueprint.bootstrap

touch "${BUILDDIR}/.out-dir"

if [ ! -z "$WRAPPER" ]; then
    cp $WRAPPER $BUILDDIR/
fi
