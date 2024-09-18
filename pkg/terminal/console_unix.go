//go:build (darwin || linux) && (amd64 || arm64)

package terminal

// SetConsole for non-windows environments is a no-op.
func SetConsole() error {
	return nil
}
