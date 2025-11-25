class Codemap < Formula
  desc "Generates a compact, visually structured 'brain map' of your codebase for LLM context"
  homepage "https://github.com/JordanCoin/codemap"
  url "https://github.com/JordanCoin/codemap/archive/refs/tags/v1.8.tar.gz"
  sha256 "UPDATE_THIS_AFTER_RELEASE"
  license "MIT"

  depends_on "go" => :build

  def install
    # Install tree-sitter queries for --deps mode
    (libexec/"queries").install Dir["scanner/queries/*.scm"]

    # Build grammars for --deps mode
    (libexec/"grammars").mkpath
    cd "scanner" do
      system "bash", "build-grammars.sh"
    end
    (libexec/"grammars").install Dir["scanner/grammars/*.dylib"]
    (libexec/"grammars").install Dir["scanner/grammars/*.so"]

    # Build from root with embedded paths
    system "go", "build", "-o", libexec/"codemap", "."

    # Create wrapper script with environment variables
    (bin/"codemap").write <<~EOS
      #!/bin/bash
      export CODEMAP_GRAMMAR_DIR="#{libexec}/grammars"
      export CODEMAP_QUERY_DIR="#{libexec}/queries"
      exec "#{libexec}/codemap" "$@"
    EOS
  end

  test do
    system "#{bin}/codemap", "."
  end
end
