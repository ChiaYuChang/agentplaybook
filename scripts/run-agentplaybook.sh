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

# --build and --update are runner flags and must not reach the CLI.
force_build=0
force_update=0
while [ "${1:-}" = "--build" ] || [ "${1:-}" = "--update" ]; do
	case "$1" in
		--build)
			force_build=1
			;;
		--update)
			force_update=1
			;;
	esac
	shift
done
if [ "${AGENTPLAYBOOK_BUILD:-}" = "1" ]; then
	force_build=1
fi
if [ "${AGENTPLAYBOOK_UPDATE:-}" = "1" ]; then
	force_update=1
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

release_os=""
case "$(uname -s 2>/dev/null || true)" in
	Linux|linux)
		release_os="linux"
		;;
	Darwin|darwin)
		release_os="darwin"
		;;
esac

release_arch=""
case "$(uname -m 2>/dev/null || true)" in
	x86_64|amd64)
		release_arch="amd64"
		;;
	aarch64|arm64)
		release_arch="arm64"
		;;
esac

platform_supported=1
if [ -z "${release_os}" ] || [ -z "${release_arch}" ]; then
	platform_supported=0
fi

if [ "${force_build}" -eq 0 ] && [ "${force_update}" -eq 0 ] && [ "${platform_supported}" -eq 1 ] && [ -x "${binary_path}" ]; then
	exec "${binary_path}" "$@"
fi

tmp_dir="${install_dir}.tmp.$$"
trap 'rm -rf "${tmp_dir}"' EXIT HUP INT TERM

prepare_tmp_dir() {
	mkdir -p "${install_dir}"
	rm -rf "${tmp_dir}"
	mkdir -p "${tmp_dir}"
}

cleanup_tmp_dir() {
	rm -rf "${tmp_dir}"
}

build_local() {
	if ! command -v go >/dev/null 2>&1; then
		echo "agentplaybook: Go toolchain is required to build agentplaybook@${cli_version}" >&2
		return 1
	fi

	prepare_tmp_dir
	(
		cd "${skill_dir}"
		GOFLAGS= GOWORK=off CGO_ENABLED=0 go build -ldflags "-X main.version=${cli_version}" -o "${tmp_dir}/${binary_name}" .
	)
	mv -f "${tmp_dir}/${binary_name}" "${binary_path}"
}

download_file() {
	url="$1"
	destination="$2"
	if command -v curl >/dev/null 2>&1; then
		curl -fL --silent --show-error "${url}" -o "${destination}"
	elif command -v wget >/dev/null 2>&1; then
		wget -q "${url}" -O "${destination}"
	else
		return 1
	fi
}

sha256_file() {
	file="$1"
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "${file}" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "${file}" | awk '{print $1}'
	else
		return 1
	fi
}

download_prebuilt() {
	archive_name="${binary_name}-${release_os}-${release_arch}.tar.gz"
	archive_path="${tmp_dir}/${archive_name}"
	checksums_path="${tmp_dir}/checksums.txt"
	record_path="${tmp_dir}/checksum.record"

	prepare_tmp_dir
	if ! download_file "${release_base_url}/${archive_name}" "${archive_path}"; then
		return 1
	fi
	if ! download_file "${release_base_url}/checksums.txt" "${checksums_path}"; then
		return 1
	fi

	awk -v target="${archive_name}" '$2 == target {print $1}' "${checksums_path}" > "${record_path}"
	record_count="$(awk 'NF {count++} END {print count + 0}' "${record_path}")"
	if [ "${record_count}" -ne 1 ]; then
		return 1
	fi

	expected_hash="$(awk 'NF {print $1}' "${record_path}")"
	if ! LC_ALL=C awk 'length($0) == 64 && $0 !~ /[^[:xdigit:]]/ {valid=1} END {exit !valid}' <<EOF
${expected_hash}
EOF
	then
		return 1
	fi
	if ! actual_hash="$(sha256_file "${archive_path}")"; then
		return 1
	fi
	expected_hash="$(printf '%s' "${expected_hash}" | tr '[:upper:]' '[:lower:]')"
	actual_hash="$(printf '%s' "${actual_hash}" | tr '[:upper:]' '[:lower:]')"
	if [ "${expected_hash}" != "${actual_hash}" ]; then
		return 1
	fi

	if ! tar -xzf "${archive_path}" -C "${tmp_dir}" "${binary_name}"; then
		return 1
	fi
	if [ -f "${tmp_dir}/${binary_name}" ] && [ ! -L "${tmp_dir}/${binary_name}" ] && [ -x "${tmp_dir}/${binary_name}" ]; then
		:
	else
		return 1
	fi
	mv -f "${tmp_dir}/${binary_name}" "${binary_path}"
}

release_base_url="${AGENTPLAYBOOK_RELEASE_BASE_URL:-https://github.com/ChiaYuChang/agentplaybook/releases/download/${cli_version}}"
release_base_url="${release_base_url%/}"

if [ "${force_build}" -eq 0 ] && [ "${platform_supported}" -eq 1 ]; then
	if download_prebuilt; then
		cleanup_tmp_dir
		exec "${binary_path}" "$@"
	fi
	cleanup_tmp_dir
fi

build_local
cleanup_tmp_dir
exec "${binary_path}" "$@"
