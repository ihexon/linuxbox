//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package port

import (
	"fmt"
	"net"
	"time"

	"bauklotze/pkg/machine/define"
)

const tcp = "tcp"

// GetFree returns a free TCP port
func GetFree() (int, error) {
	l, err := net.Listen(tcp, fmt.Sprintf("%s:%d", define.LocalHostURL, 0))
	if err != nil {
		return 0, fmt.Errorf("unable to get free TCP port: %w", err)
	}
	defer l.Close() //nolint:errcheck

	return l.Addr().(*net.TCPAddr).Port, nil
}

const defaultDialTimeout = 30 * time.Millisecond

// IsListening test someone listens to the target port
func IsListening(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", define.LocalHostURL, port), defaultDialTimeout)
	if err != nil {
		return false
	}

	_ = conn.Close()

	return true
}
