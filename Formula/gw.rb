class Gw < Formula
  desc "Git worktree manager with fzf integration"
  homepage "https://github.com/mb6611/gw"
  license "MIT"
  head "https://github.com/mb6611/gw.git", branch: "main"

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
