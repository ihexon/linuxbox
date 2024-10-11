package machine

import "testing"

func TestSpliter(t *testing.T) {
	str, ver := SplitField("dirstr/dir2@/dir3/@dir4@123/bootab@le.tar.xz@@@@v1.0.0")
	t.Logf("File: %s", str)
	t.Logf("Version: %s", ver)
}
