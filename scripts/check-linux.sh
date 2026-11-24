#!/usr/bin/env bash
# Copyright 2026 Ibrahim Karakira
# SPDX-License-Identifier: Apache-2.0
#
# Runs scripts/check.sh on Linux inside a digest-pinned golang image.
set -euo pipefail
cd "$(dirname "$0")/.."

# golang:1.27.4-bookworm
IMAGE="golang@sha256:ce59b41415e2524ae99926fe1f37ce48ba8c194dcbd2a56a65436cdcefa2bd34"

# MSYS_NO_PATHCONV stops Git Bash on Windows from rewriting the container paths.
MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd):/src" -w /src "$IMAGE" bash scripts/check.sh
