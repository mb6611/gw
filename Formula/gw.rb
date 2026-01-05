class Gw < Formula
  desc "Git worktree manager with fzf integration"
  homepage "https://github.com/mb6611/gw"
  url "https://github.com/mb6611/gw/archive/refs/tags/v0.1.1.tar.gz"
  sha256 "366c1e0274900f8e546f906bc6328527906d29f4747d3df006015e6aac93d5ee"
  license "MIT"

  depends_on "go" => :build
  depends_on "fzf"

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/gw"
  end

  def caveats
    <<~EOS
      Add to your shell config:

        eval "$(gw init zsh)"
    EOS
  end

  test do
    system "#{bin}/gw", "--help"
  end
end
