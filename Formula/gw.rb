class Gw < Formula
  desc "Git worktree manager with fzf integration"
  homepage "https://github.com/mb6611/gw"
  url "https://github.com/mb6611/gw/archive/refs/tags/v0.1.2.tar.gz"
  sha256 "6dc66b61f6b367e25985160543f392b3aaf848b864ee96fce952f7e34b1857e7"
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
