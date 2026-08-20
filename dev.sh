#!/usr/bin/env bash
# Helper script for developing svault. Usage: ./dev.sh <command>
#
#   ./dev.sh build       build binary ./svault
#   ./dev.sh run ...     build and run svault with arguments
#                        example: ./dev.sh run list
#   ./dev.sh install     build and install to /usr/local/bin (requires sudo)
#   ./dev.sh test        run all tests
#   ./dev.sh race        run tests with race detector
#   ./dev.sh cov         run tests with coverage report per package
#   ./dev.sh vet         run go vet
#   ./dev.sh fmt         format code (gofmt -w)
#   ./dev.sh check       fmt + vet + race (pre-commit check)
#   ./dev.sh dist        cross-compile for all supported platforms into dist/
#   ./dev.sh deb         build .deb and .rpm packages (requires nfpm)
#   ./dev.sh clean       remove build binary and dist/
#   ./dev.sh help        display this help message

set -euo pipefail
cd "$(dirname "$0")"

say() { printf '\033[1;36m=> %s\033[0m\n' "$*"; }

cmd="${1:-build}"
shift || true

case "$cmd" in
	build | b)
		say "make build"
		make build
		say "built: ./svault"
		;;
	run | r)
		make build >/dev/null
		say "./svault $*"
		./svault "$@"
		;;
	install | i)
		say "sudo make install"
		sudo make install
		;;
	test | t)
		say "go test ./..."
		go test ./...
		;;
	race)
		say "go test -race ./..."
		go test -race ./...
		;;
	cov)
		say "go test -cover ./..."
		go test ./... -cover
		;;
	vet)
		say "go vet ./..."
		go vet ./...
		;;
	fmt)
		say "gofmt -w ."
		gofmt -w .
		say "formatting complete"
		;;
	check)
		say "gofmt -w ."
		gofmt -w .
		say "go vet ./..."
		go vet ./...
		say "go test -race ./..."
		go test -race ./...
		say "all checks passed, ready to commit"
		;;
	dist)
		say "make dist"
		make dist
		say "artifacts saved to dist/"
		;;
	deb)
		command -v nfpm >/dev/null || {
			echo "nfpm is not installed. Run:" >&2
			echo "  go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest" >&2
			exit 1
		}
		say "make deb rpm"
		make deb rpm
		say "packages saved to dist/"
		;;
	clean)
		say "make clean"
		make clean
		;;
	help | -h | --help)
		grep '^#' "$0" | grep -v '^#!' | sed 's/^# \{0,1\}//'
		;;
	*)
		echo "unknown command: $cmd" >&2
		echo "run: ./dev.sh help" >&2
		exit 1
		;;
esac
