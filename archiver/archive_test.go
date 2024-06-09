package archiver

import "testing"

func TestArchive(t *testing.T) {
	err := DecompressFile("C:\\Users\\localuser\\tmpdir\\ovm-win.tar.zst", "C:\\Users\\localuser\\tmpdir\\ovm-win.tar", true)
	if err != nil {
		t.Fatalf("archiver.TestArchive err:%v", err)
		return
	}
}
