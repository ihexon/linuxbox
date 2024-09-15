package wsl2v2

import (
	"bauklotze/pkg/machine/wsl2v2/backend"
	"bauklotze/pkg/machine/wsl2v2/windows"
	"context"
)

func selectBackend(ctx context.Context) backend.Backend {
	return windows.Backend{
		Version: 2, // Wsl2V2
	}
}
