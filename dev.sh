#!/usr/bin/env bash
# Development script for svault.
#
# Commands:
#   ./dev.sh build       Build executable binary ./svault
#   ./dev.sh run ...     Build and execute svault with arguments
#   ./dev.sh install     Build and install to system bin directory
#   ./dev.sh test        Execute unit tests
#   ./dev.sh race        Execute unit tests with race detector
#   ./dev.sh cov         Execute unit tests with coverage reporting
#   ./dev.sh vet         Run go vet static analysis
#   ./dev.sh fmt         Format code with gofmt
#   ./dev.sh tidy        Run go mod tidy
#   ./dev.sh check       Run pre-commit checks (fmt, vet, race test)
#   ./dev.sh dist        Cross-compile binaries into dist/
#   ./dev.sh deb         Build Debian and RPM packages using nfpm
#   ./dev.sh clean       Remove build artifacts and binaries
#   ./dev.sh help        Display usage information

set -euo pipefail
cd "$(dirname "$0")"

say() {
	printf '\033[1;36m=> %s\033[0m\n' "$*"
}

cmd="${1:-build}"
if [ $# -gt 0 ]; then
	shift
fi

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
tidy)
	say "go mod tidy"
	go mod tidy
	;;
check)
	say "gofmt -w ."
	gofmt -w .
	say "go vet ./..."
	go vet ./...
	say "go test -race ./..."
	go test -race ./...
	say "all checks passed"
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
