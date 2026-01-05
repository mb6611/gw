class Gw < Formula
  desc "Git worktree manager with fzf integration"
  homepage "https://github.com/mb6611/gw"
  url "https://github.com/mb6611/gw/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "06d0055577ddf0994f6624edce463b4cc7df452f6c820a725543039515a6bda9"
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
