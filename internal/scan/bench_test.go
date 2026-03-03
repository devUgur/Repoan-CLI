package scan

import (
	"os"
	"testing"
)

func BenchmarkScan(b *testing.B) {
	// Use current directory for benchmark
	cwd, _ := os.Getwd()
	opts := ScanOptions{
		Root:             cwd,
		RespectGitignore: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Scan(opts)
	}
}
