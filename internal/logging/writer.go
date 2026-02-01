package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// RotatingWriter implements io.Writer with size-based log rotation
type RotatingWriter struct {
	path     string
	maxSize  int64
	maxFiles int

	mu   sync.Mutex
	file *os.File
	size int64
}

// NewRotatingWriter creates a new rotating writer
// It opens the log file (creating it and parent directories if needed)
func NewRotatingWriter(path string, maxSize int64, maxFiles int) (*RotatingWriter, error) {
	w := &RotatingWriter{
		path:     path,
		maxSize:  maxSize,
		maxFiles: maxFiles,
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open or create the log file
	if err := w.openFile(); err != nil {
		return nil, err
	}

	return w, nil
}

// openFile opens the log file for appending
func (w *RotatingWriter) openFile() error {
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Get current file size
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	w.file = f
	w.size = info.Size()
	return nil
}

// Write implements io.Writer
func (w *RotatingWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if rotation is needed
	if w.size+int64(len(p)) > w.maxSize {
		// If rotation fails, try to continue writing anyway
		// This is better than losing log entries
		_ = w.rotate()
	}

	n, err = w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// rotate performs log file rotation
// smoke.log -> smoke.log.1 -> smoke.log.2 -> ... -> smoke.log.N (deleted)
func (w *RotatingWriter) rotate() error {
	// Close current file
	if w.file != nil {
		_ = w.file.Close()
	}

	// Delete the oldest file if it exists
	oldest := fmt.Sprintf("%s.%d", w.path, w.maxFiles)
	_ = os.Remove(oldest) // Ignore error - file may not exist

	// Shift existing rotated files: .log.4 -> .log.5, .log.3 -> .log.4, etc.
	for i := w.maxFiles - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.path, i)
		newPath := fmt.Sprintf("%s.%d", w.path, i+1)
		_ = os.Rename(oldPath, newPath) // Ignore error - file may not exist
	}

	// Rename current log to .log.1
	_ = os.Rename(w.path, w.path+".1") // Ignore error - file may not exist

	// Open new log file
	return w.openFile()
}

// Close closes the log file
func (w *RotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}

// Sync flushes the log file to disk
func (w *RotatingWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}
