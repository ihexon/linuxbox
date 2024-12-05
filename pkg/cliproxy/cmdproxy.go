//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package cmdproxy

import (
	"fmt"

	"bauklotze/pkg/cliproxy/internal/backend"
)

func RunCMDProxy() error {
	err := backend.SSHD()
	if err != nil {
		return err //nolint:wrapcheck
	}
	return fmt.Errorf("CMDProxy running failed, %w", err)
}
