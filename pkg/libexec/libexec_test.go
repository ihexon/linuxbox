//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package libexec_test

import (
	"path/filepath"
	"testing"

	"bauklotze/pkg/libexec"
	"bauklotze/tests/fixtures"

	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	dir := fixtures.GetTestFixtures("libexec")

	err := libexec.Setup(filepath.Join(dir, "bin", "ovm"))
	require.NoError(t, err)
}

func TestFindBinary(t *testing.T) {
	dir := fixtures.GetTestFixtures("libexec")

	err := libexec.Setup(filepath.Join(dir, "bin", "ovm"))
	require.NoError(t, err)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "gvproxy",
			wantErr: false,
		},
		{
			name:    "krunkit",
			wantErr: false,
		},
		{
			name:    "ovm",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := libexec.FindBinary(tt.name)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDYLDLibraryPath(t *testing.T) {
	dir := fixtures.GetTestFixtures("libexec")

	err := libexec.Setup(filepath.Join(dir, "bin", "ovm"))
	require.NoError(t, err)

	require.Equal(t, filepath.Join(dir, "libexec"), libexec.GetDYLDLibraryPath())
}
