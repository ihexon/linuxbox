package archiver

import (
	"bauklotze/pkg/archiver/decompress"
	"bauklotze/pkg/machine/define"
	"github.com/mholt/archiver/v4"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchive(t *testing.T) {
	cwd, _ := os.Getwd()
	t.Logf(cwd)

}

func TestIdentify(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		content    []byte
		wantFormat string
	}{
		{
			name:       "xz file",
			filename:   "test_file/test_data.tar.xz",
			wantFormat: ".xz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.OpenFile(tt.filename, os.O_RDONLY, 644)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			format, reader, err := Identify(tt.filename, file)
			if err != nil {
				t.Fatalf("Identify() error = %v", err)
			}
			if format.Name() != tt.wantFormat {
				t.Errorf("Identify() got = %v, want %v", format.Name(), tt.wantFormat)
			}

			decom, ok := format.(archiver.Decompressor)
			if ok {
				c, err := decom.OpenReader(reader)
				if err != nil {
					t.Fatal(err)
				}
				defer c.Close()
				outputDir := filepath.Join("test_file", "decompress_dir")
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatal(err)
				}
				decompressedFilename := strings.TrimSuffix(filepath.Base(tt.filename), ".xz")
				outputFile, err := os.Create(filepath.Join(outputDir, decompressedFilename))
				if err != nil {
					t.Fatal(err)
				}
				defer outputFile.Close()
				if wed, err := io.Copy(outputFile, c); err != nil {
					t.Fatalf("Failed to decompress: %v", err)
				} else {
					t.Logf("Decompressed file: %d bytes", wed)
				}
			}
		})
	}
}

func TestDecompress(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		content    []byte
		wantFormat string
	}{
		{
			name:       "xz file",
			filename:   "test_file/test_data.tar.xz",
			wantFormat: ".xz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &define.VMFile{
				Path: tt.filename,
			}
			// Wed Aug  7 04:33:19 PM HKT 2024
			// This should be ok for now :)
			err := decompress.Decompress(f, "test_file/decompress_dir")
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
