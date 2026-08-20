VERSION := $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev")
LDFLAGS  := -ldflags "-X 'svault/cmd.version=$(VERSION)' -s -w"

PREFIX ?= /usr/local

.PHONY: build dist test vet clean install uninstall deb rpm

build:
	go build $(LDFLAGS) -o svault .

install: build
	install -m 0755 svault $(PREFIX)/bin/svault

uninstall:
	rm -f $(PREFIX)/bin/svault

dist:
	mkdir -p dist
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/svault-linux-amd64       .
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o dist/svault-linux-arm64       .
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/svault-darwin-amd64      .
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/svault-darwin-arm64      .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/svault-windows-amd64.exe .
	cd dist && sha256sum * > checksums.txt

# Stage the per-arch binary at a fixed path (nfpm does not expand env vars in
# contents.src) and build the package. $(1)=arch $(2)=format
define nfpm_build
	cp dist/svault-linux-$(1) dist/svault.pkgbin
	VERSION=$(VERSION) ARCH=$(1) nfpm package -f packaging/nfpm.yaml -p $(2) -t dist/
	rm -f dist/svault.pkgbin
endef

# Build .deb packages for amd64 and arm64. Requires `dist` first and nfpm
# (https://nfpm.goreleaser.com). Output goes to dist/.
deb: dist
	$(call nfpm_build,amd64,deb)
	$(call nfpm_build,arm64,deb)

# Build .rpm packages for amd64 and arm64.
rpm: dist
	$(call nfpm_build,amd64,rpm)
	$(call nfpm_build,arm64,rpm)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf dist/ svault
