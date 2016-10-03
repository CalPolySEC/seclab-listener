package backend

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// Backend is the interface used by the frontend to handle Open/Close events
type Backend interface {
	Open() error
	Close() error
}

type fileBackend struct {
	linkPath   string
	openPath   string
	closedPath string
}

// New creates a new instance of a Backend
func New(linkPath, openPath, closedPath string) Backend {
	return &fileBackend{
		linkPath:   linkPath,
		openPath:   openPath,
		closedPath: closedPath,
	}
}

// Atomically hardlink src to dst, overwriting dst
// This is achieved through a hardlink to a temporary file, followed by a move
// It is assumed that src and dst are on the same device
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
