package backend

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type Backend interface {
	Open() error
	Close() error
}

type fileBackend struct {
	linkPath   string
	openPath   string
	closedPath string
}

func New(linkPath, openPath, closedPath string) Backend {
	return &fileBackend{
		linkPath:   linkPath,
		openPath:   openPath,
		closedPath: closedPath,
	}
}

// Atomically point a
func atomicLink(src, dst string) error {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "templink")
	if err := os.Link(src, tempFile); err != nil {
		return err
	}
	return os.Rename(tempFile, dst)
}

func (b *fileBackend) Open() error {
	return atomicLink(b.openPath, b.linkPath)
}

func (b *fileBackend) Close() error {
	return atomicLink(b.closedPath, b.linkPath)
}
