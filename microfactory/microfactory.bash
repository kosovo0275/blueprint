build_go() {
    local mf_version=3
    local mf_src="${BLUEPRINTDIR}/microfactory"
    local mf_bin="${BUILDDIR}/microfactory_$(uname)"
    local mf_version_file="${BUILDDIR}/.microfactory_$(uname)_version"
    local built_bin="${BUILDDIR}/$1"
    local from_src=1

    if [ -f "${mf_bin}" ] && [ -f "${mf_version_file}" ]; then
        if [ "${mf_version}" -eq "$(cat "${mf_version_file}")" ]; then
            from_src=0
        fi
    fi

    local mf_cmd
    if [ $from_src -eq 1 ]; then
        local gen_src_dir="${BUILDDIR}/.microfactory_$(uname)_intermediates/src"
        mkdir -p "${gen_src_dir}"
        sed "s/^package microfactory/package main/" "${mf_src}/microfactory.go" >"${gen_src_dir}/microfactory.go"
        mf_cmd="go run -buildmode pie ${gen_src_dir}/microfactory.go"
    else
        mf_cmd="${mf_bin}"
    fi

    rm -f "${BUILDDIR}/.$1.trace"
    GOROOT=$(builtin cd $GOROOT; pwd) ${mf_cmd} -b "${mf_bin}" \
            -pkg-path "github.com/google/blueprint=${BLUEPRINTDIR}" \
            -trimpath "${SRCDIR}" \
            ${EXTRA_ARGS} \
            -o "${built_bin}" $2

    if [ $? -eq 0 ] && [ $from_src -eq 1 ]; then
        echo "${mf_version}" >"${mf_version_file}"
    fi
}
