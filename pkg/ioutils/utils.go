package ioutils

import (
	"fmt"
	"os"
)

// Exists checks whether a file or directory exists at the given path.
func Exists(path string) error {
	_, err := os.Stat(path)
	return err
}

func BExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

// Lexists checks whether a file or directory exists at the given path, without
// resolving symlinks
func Lexists(path string) error {
	_, err := os.Lstat(path)
	return err
}
func GuardedRemoveAll(path string) error {
	if path == "" || path == "/" {
		return fmt.Errorf("refusing to recursively delete `%s`", path)
	}
	return os.RemoveAll(path)
}
