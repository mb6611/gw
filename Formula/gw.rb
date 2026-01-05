class Gw < Formula
  desc "Git worktree manager with fzf integration"
  homepage "https://github.com/mb6611/gw"
  url "https://github.com/mb6611/gw/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "PLACEHOLDER"
  license "MIT"

  depends_on "go" => :build
  depends_on "fzf"

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/gw"
  end

  def caveats
    <<~EOS
      To enable directory switching, add this to your shell config:

      # ~/.zshrc
      eval "$(gw init zsh)"

      # ~/.bashrc
      eval "$(gw init bash)"

      # ~/.config/fish/config.fish
      gw init fish | source
    EOS
  end

  test do
    system "#{bin}/gw", "--help"
  end
end
