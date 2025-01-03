//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package port_test

import (
	"net"
	"net/http/httptest"
	"testing"

	"bauklotze/pkg/port"

	"github.com/stretchr/testify/require"
)

func TestPort(t *testing.T) {
	s := httptest.NewServer(nil)

	p, err := port.GetFree(s.Listener.Addr().(*net.TCPAddr).Port)
	require.NoError(t, err)
	require.NotEqual(t, s.Listener.Addr().(*net.TCPAddr).Port, p)

	p, err = port.GetFree(61212)
	require.NoError(t, err)
	require.Equal(t, 61212, p)
}

func TestIsListening(t *testing.T) {
	s := httptest.NewServer(nil)

	require.True(t, port.IsListening(s.Listener.Addr().(*net.TCPAddr).Port))
	require.False(t, port.IsListening(0))
}
