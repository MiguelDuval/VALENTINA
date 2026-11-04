#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
expected_commit="f01c1d84c08b1de972d8cc268da25602598adff8"

cd "$repo_root"
echo "Synchronizing and initializing VALENTINA upstream submodules..."
git submodule sync --recursive
git submodule update --init --recursive

actual_commit="$(git -C upstream/huginn rev-parse HEAD)"
if [[ "$actual_commit" != "$expected_commit" ]]; then
  echo "Unexpected Huginn revision. Expected $expected_commit, got $actual_commit." >&2
  exit 1
fi

echo "Huginn source is ready at upstream/huginn."
echo "Pinned revision: $actual_commit"
