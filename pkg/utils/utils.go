package utils

import (
	"fmt"
	"os"
)

func GuardedRemoveAll(path string) error {
	if path == "" || path == "/" || path == "/dev" || path == "/proc" || path == "/sys" {
		return fmt.Errorf("refusing to recursively delete `%s`", path)
	}
	return os.RemoveAll(path)
}
