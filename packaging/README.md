# Packaging and Distribution

How to publish svault to the major package managers. Each section is independent.

## AUR (Arch Linux)

Two packages are provided:

- `packaging/aur/PKGBUILD` builds from source (package name `svault`)
- `packaging/aur-bin/PKGBUILD` installs the prebuilt binary (package name `svault-bin`)

### One-time setup

1. Create an account at <https://aur.archlinux.org> and add your SSH public key
   under My Account.
2. Check the name is free: <https://aur.archlinux.org/packages/svault>. If taken,
   rename the package (for example `svault-vault`).

### Publish

```bash
# Clone the (empty) AUR repo for the package
git clone ssh://aur@aur.archlinux.org/svault.git aur-svault
cd aur-svault

# Copy the PKGBUILD in
cp ../svault/packaging/aur/PKGBUILD .

# Generate the metadata file AUR requires
makepkg --printsrcinfo > .SRCINFO

# Test that it builds and installs cleanly before pushing
makepkg -si

git add PKGBUILD .SRCINFO
git commit -m "Initial import: svault 2.0.1"
git push
```

### Update on a new release

```bash
# bump pkgver in PKGBUILD, reset pkgrel to 1
updpkgsums          # only needed if sha256sums are not SKIP
makepkg --printsrcinfo > .SRCINFO
git commit -am "Update to 2.0.1"
git push
```

## Homebrew

Published as a tap at <https://github.com/dafagareth/homebrew-tap>. The canonical
formula in this repo is `Formula/svault.rb`; the tap repo holds a copy under
`Formula/svault.rb`.

Users install with:

```bash
brew install dafagareth/tap/svault
```

### Update on a new release

1. In `Formula/svault.rb`, bump `version` and replace each `sha256` from the
   release `checksums.txt` (match by file name).
2. Copy the formula into the tap repo and push:

   ```bash
   cp Formula/svault.rb ../homebrew-tap/Formula/svault.rb
   cd ../homebrew-tap && git commit -am "svault <version>" && git push
   ```

Submitting to `homebrew-core` (so `brew install svault` works without a tap) has
a high bar: the project generally needs to be notable (stars, forks, watchers)
and stable. A tap is the right starting point.

## Scoop (Windows)

This repo doubles as a Scoop bucket: the manifest lives at `bucket/svault.json`,
which Scoop discovers automatically. No separate bucket repo is needed.

Users install with:

```powershell
scoop bucket add svault https://github.com/dafagareth/svault
scoop install svault
```

### Update on a new release

In `bucket/svault.json`, bump `version`, the download `url`, and the `hash`
(from the release `checksums.txt`, file `svault-windows-amd64.exe`), then commit
and push. The manifest's `autoupdate`/`checkver` blocks let `scoop update` pick
up future versions automatically.

## Nix

Add a flake or a derivation that uses `buildGoModule`:

```nix
buildGoModule {
  pname = "svault";
  version = "2.0.0";
  src = fetchFromGitHub {
    owner = "dafagareth";
    repo = "svault";
    rev = "v2.0.0";
    hash = "REPLACE_WITH_HASH";
  };
  vendorHash = "REPLACE_WITH_VENDOR_HASH";
  ldflags = [ "-s" "-w" "-X" "svault/cmd.version=2.0.0" ];
}
```

Submitting to `nixpkgs` is a PR against <https://github.com/NixOS/nixpkgs>.

## Debian / Ubuntu and Fedora (.deb / .rpm)

The simplest way to support Debian, Ubuntu, and Fedora users is to ship `.deb`
and `.rpm` files on the GitHub release. We use [nfpm](https://nfpm.goreleaser.com),
which turns one YAML config (`packaging/nfpm.yaml`) into both formats without the
full Debian toolchain.

Install nfpm once:

```bash
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
```

Build the packages (after the binaries exist):

```bash
make deb    # produces dist/svault_<version>_amd64.deb and _arm64.deb
make rpm    # produces dist/svault-<version>.x86_64.rpm and aarch64.rpm
```

Users then install with:

```bash
sudo dpkg -i svault_2.0.0_amd64.deb      # Debian / Ubuntu
sudo rpm -i svault-2.0.0.x86_64.rpm      # Fedora / RHEL
```

The release workflow builds and attaches these automatically on every tag.

## Official Debian repository

Getting into the official Debian archive (so plain `apt install svault` works) is
a separate, much heavier process:

1. File an ITP (Intent To Package) bug against the `wnpp` pseudo-package.
2. Create a Debian-policy-compliant `debian/` directory.
3. Find a sponsor (a Debian Developer) to review and upload it.
4. Maintain it through Debian's release cycle.

This generally expects an established, maintained project and takes months. The
self-hosted `.deb` above is the practical path until then. The same applies to
Fedora's and Arch's official repositories.
