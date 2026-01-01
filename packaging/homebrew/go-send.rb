class GoSend < Formula
  desc "Secure, end-to-end encrypted file sharing with store-and-forward"
  homepage "https://github.com/VinMeld/go-send"
  url "https://github.com/VinMeld/go-send/archive/refs/tags/v2.1.0.tar.gz"
  sha256 "5519911e8b5a76ba3e1c76fc96d4f139daf4d5ab939ece29a0006684eeed743f"
  license "GPL-3.0"
  head "https://github.com/VinMeld/go-send.git", branch: "main"

  depends_on "go" => :build

  def install
    # Build the client binary
    system "go", "build", *std_go_args(output: bin/"go-send", ldflags: "-s -w"), "./cmd/client"
    
    # Optionally build the server binary (uncomment if you want to include it)
    # system "go", "build", *std_go_args(output: bin/"go-send-server", ldflags: "-s -w"), "./cmd/server"
  end

  test do
    # Test that the binary runs and shows version/help
    assert_match "go-send", shell_output("#{bin}/go-send --help 2>&1")
  end

  def caveats
    <<~EOS
      To get started with go-send:
        1. Initialize your configuration: go-send config init --user <username>
        2. Set your server URL: go-send set-server <server-url>
        3. Register with server: go-send register --token <token>
        
      For more information, visit: https://github.com/VinMeld/go-send
    EOS
  end
end
