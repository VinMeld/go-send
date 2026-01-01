class GoSend < Formula
  desc "Secure, end-to-end encrypted file sharing with store-and-forward"
  homepage "https://github.com/VinMeld/go-send"
  version "2.1.0"
  license "GPL-3.0"

  on_macos do
    on_arm do
      url "https://github.com/VinMeld/go-send/releases/download/v2.1.0/go-send_2.1.0_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
    end
    on_intel do
      url "https://github.com/VinMeld/go-send/releases/download/v2.1.0/go-send_2.1.0_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/VinMeld/go-send/releases/download/v2.1.0/go-send_2.1.0_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
    end
    on_intel do
      url "https://github.com/VinMeld/go-send/releases/download/v2.1.0/go-send_2.1.0_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
    end
  end

  def install
    bin.install "go-send"
  end

  test do
    assert_match "go-send", shell_output("#{bin}/go-send --help 2>&1")
  end

  def caveats
    <<~EOS
      To get started with go-send:
        1. Initialize: go-send config init --user <username>
        2. Set server: go-send set-server <server-url>
        3. Register: go-send register --token <token>
        
      Documentation: https://github.com/VinMeld/go-send
    EOS
  end
end
