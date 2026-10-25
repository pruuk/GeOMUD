#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/../.." && pwd)"

ACT_FLAGS="${ACT_FLAGS:---pull=false -P ubuntu-24.04=catthehacker/ubuntu:act-latest}"
XDG_CONFIG_HOME="${ACT_CONFIG_HOME:-${repo_root}/.github}"
export XDG_CONFIG_HOME
read -r -a act_flags <<<"$ACT_FLAGS"

mkdir -p "${XDG_CONFIG_HOME}/act"
touch "${XDG_CONFIG_HOME}/act/actrc"

run_act() {
	local event="$1"
	local event_file="$2"
	local workflow="$3"
	shift 3

	act "${act_flags[@]}" --dryrun "$event" "$@" \
		-e "$event_file" \
		-W "$workflow"
}

# CI combines the old lint and PR test workflows. Dry-run both event shapes
# because pull requests cancel superseded runs while pushes to master do not.
run_act push .github/act/push_master.json .github/workflows/ci.yml
run_act pull_request .github/act/pull_request.json .github/workflows/ci.yml
