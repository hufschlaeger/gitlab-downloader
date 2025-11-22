package http

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestFileAdapter_CreateFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.bin")

	a := NewFileAdapter()
	wc, err := a.CreateFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// write some bytes
	if _, err := io.WriteString(wc, "hello"); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	// Close once and verify
	if err := wc.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if string(b) != "hello" {
		t.Fatalf("unexpected content: %q", string(b))
	}
}

func TestFileAdapter_CreateFile_Error(t *testing.T) {
	// point to a non-existent subdir; os.Create will fail because parent does not exist
	dir := t.TempDir()
	badPath := filepath.Join(dir, "missing", "out.bin")
	a := NewFileAdapter()
	_, err := a.CreateFile(badPath)
	if err == nil {
		t.Fatalf("expected error for missing parent directory")
	}
	if got := err.Error(); got == "" || got[:22] != "failed to create file:"[:22] {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}
