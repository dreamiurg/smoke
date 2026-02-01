package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRotatingWriter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	w, err := NewRotatingWriter(path, 1024, 3)
	if err != nil {
		t.Fatalf("NewRotatingWriter() error = %v", err)
	}
	defer w.Close()

	// File should exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("log file was not created")
	}
}

func TestNewRotatingWriterCreatesDir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subdir", "nested")
	path := filepath.Join(subdir, "test.log")

	w, err := NewRotatingWriter(path, 1024, 3)
	if err != nil {
		t.Fatalf("NewRotatingWriter() error = %v", err)
	}
	defer w.Close()

	// Directory should exist
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		t.Error("log directory was not created")
	}
}

func TestRotatingWriterWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	w, err := NewRotatingWriter(path, 1024, 3)
	if err != nil {
		t.Fatalf("NewRotatingWriter() error = %v", err)
	}
	defer w.Close()

	msg := []byte("test message\n")
	n, err := w.Write(msg)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != len(msg) {
		t.Errorf("Write() = %d, want %d", n, len(msg))
	}

	// Sync to ensure written
	w.Sync()

	// Read back
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != string(msg) {
		t.Errorf("file content = %q, want %q", string(data), string(msg))
	}
}

func TestRotatingWriterRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	// Small max size to trigger rotation
	maxSize := int64(50)
	w, err := NewRotatingWriter(path, maxSize, 3)
	if err != nil {
		t.Fatalf("NewRotatingWriter() error = %v", err)
	}
	defer w.Close()

	// Write enough data to trigger rotation
	msg := strings.Repeat("x", 30) + "\n"
	for i := 0; i < 5; i++ {
		_, err := w.Write([]byte(msg))
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}
	}
	w.Sync()

	// Check that rotation happened
	rotated1 := path + ".1"
	if _, err := os.Stat(rotated1); os.IsNotExist(err) {
		t.Error("expected rotated file .1 to exist")
	}
}

func TestRotatingWriterMaxFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	// Very small max size and only 2 rotated files
	maxSize := int64(20)
	maxFiles := 2
	w, err := NewRotatingWriter(path, maxSize, maxFiles)
	if err != nil {
		t.Fatalf("NewRotatingWriter() error = %v", err)
	}
	defer w.Close()

	// Write enough data to trigger multiple rotations
	msg := strings.Repeat("x", 15) + "\n"
	for i := 0; i < 10; i++ {
		_, err := w.Write([]byte(msg))
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}
	}
	w.Sync()

	// Should have at most maxFiles rotated files
	for i := maxFiles + 1; i <= maxFiles+3; i++ {
		rotated := filepath.Join(dir, "test.log."+string(rune('0'+i)))
		if _, err := os.Stat(rotated); !os.IsNotExist(err) {
			t.Errorf("file %s should not exist (max %d files)", rotated, maxFiles)
		}
	}
}
