#!/usr/bin/env bash
# Copyright 2026 Ibrahim Karakira
# SPDX-License-Identifier: Apache-2.0
#
# Full local check: format, vet, test, build. Exits non-zero on the first failure.
# Runs in Git Bash on Windows and in bash on Linux.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "== gofmt"
unformatted="$(gofmt -l .)"
if [ -n "$unformatted" ]; then
	echo "gofmt: these files need formatting:" >&2
	echo "$unformatted" >&2
	exit 1
fi

echo "== licence headers"
# No `|| true` here: if git fails (not a work tree, dubious ownership, git missing),
# set -e stops the script instead of passing on an empty file list.
files="$(git ls-files --cached --others --exclude-standard -- '*.go' '*.sh')"
missing=""
while IFS= read -r f; do
	[ -n "$f" ] || continue
	rc=0
	grep -q 'SPDX-License-Identifier: Apache-2.0' -- "$f" || rc=$?
	case "$rc" in
	0) ;;
	1) missing="${missing}${f}"$'\n' ;;
	*)
		echo "licence headers: could not read $f" >&2
		exit 1
		;;
	esac
done <<<"$files"
if [ -n "$missing" ]; then
	echo "these files lack an Apache-2.0 SPDX header:" >&2
	echo "$missing" >&2
	exit 1
fi

echo "== go vet"
go vet ./...

echo "== go test"
# The race detector needs cgo and a C compiler (https://go.dev/doc/articles/race_detector).
if [ "$(go env CGO_ENABLED)" = "1" ] && command -v "$(go env CC)" >/dev/null 2>&1; then
	go test -race ./...
else
	echo "note: cgo or a C compiler is not available, running tests without -race"
	go test ./...
fi

echo "== go build"
go build ./...

echo "check: ok"
