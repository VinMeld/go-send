class GoSend < Formula
  desc "Secure, end-to-end encrypted file sharing with store-and-forward"
  homepage "https://github.com/VinMeld/go-send"
  version "2.1.0"
  license "GPL-3.0"

  on_macos do
    on_arm do
      url "https://github.com/VinMeld/go-send/releases/download/v2.1.0/go-send_2.1.0_darwin_arm64.tar.gz"
      sha256 "959d355e86e20bb19b6ea0b9ca73ddb619f906dec37d51ac0b5c2714779c7dfa"
    end
    on_intel do
      url "https://github.com/VinMeld/go-send/releases/download/v2.1.0/go-send_2.1.0_darwin_amd64.tar.gz"
      sha256 "287d839b7efebe5bbe818987ae8a5a4f75ac75f41c3be01646c116c823934cd9"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/VinMeld/go-send/releases/download/v2.1.0/go-send_2.1.0_linux_arm64.tar.gz"
      sha256 "6cae52f59fa42fbee38693d4255e5f29eb0c4975af45265d35d81b35c736379c"
    end
    on_intel do
      url "https://github.com/VinMeld/go-send/releases/download/v2.1.0/go-send_2.1.0_linux_amd64.tar.gz"
      sha256 "43addc7668c78e39cceec9b57a22c7b2ba73f188a1af959a70b7a9655ce9def1"
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
