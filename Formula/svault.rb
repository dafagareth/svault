# Homebrew formula for svault, the local encrypted secret vault.
#
# This installs a prebuilt binary from the GitHub release, so it does not need
# a Go toolchain.
#
# Intended for a tap. Publish it as a repo named `homebrew-tap` under your
# account, then users install with:
#
#   brew install dafagareth/tap/svault
#
# UPDATING ON A NEW RELEASE:
#   1. Bump `version` below.
#   2. Replace each REPLACE_WITH_SHA256 with the matching value from the
#      release's checksums.txt:
#        curl -fsSL https://github.com/dafagareth/svault/releases/download/v<version>/checksums.txt
#      Match by file name (svault-darwin-arm64, svault-darwin-amd64).

class Svault < Formula
  desc "Local encrypted secret vault for developers"
  homepage "https://github.com/dafagareth/svault"
  version "1.0.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/dafagareth/svault/releases/download/v#{version}/svault-darwin-arm64"
      sha256 "b8f93caf518b61b1c90ea41d941df86a65bd731e155c88f90d3333c9c559d46e"
    end
    on_intel do
      url "https://github.com/dafagareth/svault/releases/download/v#{version}/svault-darwin-amd64"
      sha256 "1be733793d8559d6a95e3f2b822fb03f288373709604e902dc4a5d3da34530ba"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/dafagareth/svault/releases/download/v#{version}/svault-linux-arm64"
      sha256 "09fde61725184bcf16ec7a6bfeee189d63ae0cf8566b9e69e081ac9a1a6652c4"
    end
    on_intel do
      url "https://github.com/dafagareth/svault/releases/download/v#{version}/svault-linux-amd64"
      sha256 "9a2e05afa56661a1a3e3be712d5a7a91d4294770859aa1bfc87e59919b3179db"
    end
  end

  def install
    # The downloaded asset keeps its release name; install it as `svault`.
    bin.install Dir["*"].first => "svault"
  end

  test do
    assert_match "svault v#{version}", shell_output("#{bin}/svault version")
  end
end
