class Pancake < Formula
    desc "Your pancake CLI tool"
    homepage "https://github.com/a6h15hek/pancake"
    url "https://github.com/a6h15hek/pancake/archive/refs/tags/v1.1.0.tar.gz"
    sha256 "731552b3853b9549c83d24862bdf230f0c727ba84fd1f912ebf22955d3bf4046"
    license "MIT"
  
    def install
      system "GOOS=linux GOARCH=amd64 go build -o pancake-linux-amd64"
      bin.install "pancake-linux-amd64" => "pancake"
  
      if OS.mac?
        system "GOOS=darwin GOARCH=amd64 go build -o pancake-darwin-amd64"
        bin.install "pancake-darwin-amd64" => "pancake"
      elsif OS.windows?
        system "GOOS=windows GOARCH=amd64 go build -o pancake-windows-amd64.exe"
        bin.install "pancake-windows-amd64.exe" => "pancake.exe"
      end
    end
  
    test do
      assert_match "pancake version", shell_output("#{bin}/pancake version")
    end
  end
  