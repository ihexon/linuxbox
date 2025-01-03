//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package port

import (
	"fmt"
	"net"
	"time"
)

func GetFree(defaultPort int) (int, error) {
	if defaultPort != 0 && !IsListening(defaultPort) {
		return defaultPort, nil
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("unable to get free TCP port: %w", err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

const defaultDialTimeout = 30 * time.Millisecond

func IsListening(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", port), defaultDialTimeout)
	if err != nil {
		return false
	}

	_ = conn.Close()

	return true
}
