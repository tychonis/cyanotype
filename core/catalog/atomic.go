package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// atomicWrite writes data to dst atomically (POSIX-style): write to a temp
// file in the same directory, fsync, close, then rename into place.
// It also fsyncs the parent directory so the rename is durable.
func atomicWrite(dst string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }

	// Ensure cleanup if we return before successful rename.
	defer cleanup()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("fsync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	// Windows: os.Rename doesn't replace existing files—remove first if present.
	if runtime.GOOS == "windows" {
		if _, err := os.Stat(dst); err == nil {
			_ = os.Remove(dst)
		}
	}

	if err := os.Rename(tmpName, dst); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", tmpName, dst, err)
	}

	if err := os.Chmod(dst, perm); err != nil {
		return fmt.Errorf("chmod %s: %w", dst, err)
	}

	// Fsync parent dir so the rename entry is durable.
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}

	return nil
}
