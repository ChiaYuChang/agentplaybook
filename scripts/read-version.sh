#!/usr/bin/env sh
set -eu

LC_ALL=C
export LC_ALL

script_path=$0
case "$script_path" in
	/*) ;;
	*) script_path="./${script_path}" ;;
esac
script_dir="$(CDPATH= cd "$(dirname "$script_path")" && pwd -P)"
version_file="${script_dir}/../internal/version/VERSION"

fail() {
	printf 'agentplaybook: invalid canonical release version in %s: %s\n' "$version_file" "$1" >&2
	exit 1
}

if [ ! -f "$version_file" ] || [ ! -r "$version_file" ]; then
	fail "file missing or unreadable"
fi

if ! exec 3< "$version_file"; then
	fail "file cannot be opened"
fi
if ! IFS= read -r version <&3; then
	fail "expected one LF-terminated version line"
fi
extra=
if IFS= read -r extra <&3; then
	fail "expected exactly one line"
fi
if [ -n "$extra" ]; then
	fail "unexpected trailing data"
fi
exec 3<&-

actual_bytes="$(wc -c < "$version_file" | tr -d '[:space:]')"
expected_bytes=$((${#version} + 1))
if [ "$actual_bytes" != "$expected_bytes" ]; then
	fail "unexpected byte count"
fi

if ! printf '%s\n' "$version" | awk '
	$0 !~ /^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$/ { invalid = 1 }
	END { exit invalid }
'; then
	fail "expected v<major>.<minor>.<patch> without leading zeroes"
fi

printf '%s\n' "$version"
