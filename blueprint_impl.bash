if [ ! "${BLUEPRINT_BOOTSTRAP_VERSION}" -eq "2" ]; then
  echo "Please run bootstrap.bash again (out of date)" >&2
  exit 1
fi

if [ -z "$BLUEPRINT_LIST_FILE" ]; then
  OUR_LIST_FILE="${BUILDDIR}/.bootstrap/bplist"
  TEMP_LIST_FILE="${OUR_FILES_LIST}.tmp"
  mkdir -p "$(dirname ${OUR_LIST_FILE})"
  (builtin cd "$SRCDIR";
    find . -mindepth 1 -type d \( -name ".*" -o -execdir test -e {}/.out-dir \; \) -prune -o -name $TOPNAME -print | sort) >"${TEMP_LIST_FILE}"
  if cmp -s "${OUR_LIST_FILE}" "${TEMP_LIST_FILE}"; then
    rm "${TEMP_LIST_FILE}"
  else
    mv "${TEMP_LIST_FILE}" "${OUR_LIST_FILE}"
  fi
  BLUEPRINT_LIST_FILE="${OUR_LIST_FILE}"
fi

export GOROOT
export BLUEPRINT_LIST_FILE

source "${BLUEPRINTDIR}/microfactory/microfactory.bash"

BUILDDIR="${BUILDDIR}/.minibootstrap" build_go minibp github.com/google/blueprint/bootstrap/minibp

"${NINJA}" -v -f "${BUILDDIR}/.minibootstrap/build.ninja"

"${NINJA}" -v -f "${BUILDDIR}/.bootstrap/build.ninja"

if [ -z "$SKIP_NINJA" ]; then
    "${NINJA}" -v -f "${BUILDDIR}/build.ninja" "$@"
else
    exit 0
fi
