#!/usr/bin/env sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)"
skill_dir="$(CDPATH= cd -- "${script_dir}/.." && pwd -P)"

cli_version="$(tr -d '[:space:]' < "${script_dir}/VERSION")"
binary_name="agentplaybook"

if [ -n "${XDG_CACHE_HOME:-}" ]; then
	cache_root="${XDG_CACHE_HOME}/agentplaybook"
elif [ -n "${HOME:-}" ]; then
	cache_root="${HOME}/.cache/agentplaybook"
else
	echo "agentplaybook: HOME or XDG_CACHE_HOME must be set" >&2
	exit 1
fi

# AGENTPLAYBOOK_DEV=1 (or WORKFLOW_DEV=1) runs the binary built directly from the local workspace.
if [ -n "${AGENTPLAYBOOK_DEV:-}" ] || [ -n "${WORKFLOW_DEV:-}" ]; then
	dev_binary="${cache_root}/dev/${binary_name}"
	mkdir -p "${cache_root}/dev"
	(
		cd "${skill_dir}"
		GOFLAGS= GOWORK=off CGO_ENABLED=0 go build -ldflags "-X main.version=${cli_version}-dev" -o "${dev_binary}" .
	)
	exec "${dev_binary}" "$@"
fi

install_dir="${cache_root}/${cli_version}"
binary_path="${install_dir}/${binary_name}"

if [ ! -x "${binary_path}" ]; then
	if ! command -v go >/dev/null 2>&1; then
		echo "agentplaybook: Go toolchain is required to build agentplaybook@${cli_version}" >&2
		exit 1
	fi

	tmp_dir="${install_dir}.tmp.$$"
	rm -rf "${tmp_dir}"
	mkdir -p "${tmp_dir}"
	trap 'rm -rf "${tmp_dir}"' EXIT HUP INT TERM

	(
		cd "${skill_dir}"
		GOFLAGS= GOWORK=off CGO_ENABLED=0 go build -ldflags "-X main.version=${cli_version}" -o "${tmp_dir}/${binary_name}" .
	)

	mkdir -p "${install_dir}"
	mv "${tmp_dir}/${binary_name}" "${binary_path}"
	rm -rf "${tmp_dir}"
fi

exec "${binary_path}" "$@"
