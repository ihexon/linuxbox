package cmdproxy

import (
	"bauklotze/pkg/cliproxy/internal/backend"
	"fmt"
)

func RunCMDProxy() error {
	var err error
	err = backend.SSHD()
	if err != nil {
		return err
	}
	return fmt.Errorf("CMDProxy running failed, %v", err)
}
