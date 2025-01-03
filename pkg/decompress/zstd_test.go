//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package decompress_test

import (
	"os"
	"path/filepath"
	"testing"

	"bauklotze/pkg/decompress"
	"bauklotze/tests/fixtures"

	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
)

func TestZstd(t *testing.T) {
	dir := fixtures.GetTestFixtures("zstd")

	tmp := t.TempDir()
	t.Cleanup(func() {
		_ = os.RemoveAll(tmp)
	})

	src := filepath.Join(dir, "1.txt.zst")
	target := filepath.Join(tmp, "1.txt")

	err := decompress.Zstd(src, target)
	require.NoError(t, err)

	b, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(b))
}
