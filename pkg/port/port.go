//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package port

import (
	"fmt"
	"net"
	"strconv"
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

	_, randomPort, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, fmt.Errorf("unable to determine free port: %w", err)
	}

	rp, err := strconv.Atoi(randomPort)
	if err != nil {
		return 0, fmt.Errorf("unable to convert random port to int: %w", err)
	}
	return rp, nil
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
