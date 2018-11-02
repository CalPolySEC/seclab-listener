package backend

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// Backend is the interface used by the frontend to handle Open/Close events
type Backend interface {
	Open() error
	Close() error
	Coffee() error
	Fire() error
}

type fileBackend struct {
	linkPath   string
	openPath   string
	closedPath string
	coffeePath string
	firePath   string
}

// New creates a new instance of a Backend
func New(linkPath, openPath, closedPath, coffeePath, firePath string) Backend {
	return &fileBackend{
		linkPath:   linkPath,
		openPath:   openPath,
		closedPath: closedPath,
		coffeePath: coffeePath,
		firePath:   firePath,
	}
}

// Atomically hardlink src to dst, overwriting dst, and update the timestamp.
// This is achieved through a hardlink to a temporary file, followed by a move,
// since move is atomic. We're assuming that src and dst are on the same device
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

	if err := os.Rename(tempFile, dst); err != nil {
		return err
	}

	return os.Chtimes(dst, time.Now(), time.Now())
}

func (b *fileBackend) Open() error {
	return atomicLink(b.openPath, b.linkPath)
}

func (b *fileBackend) Close() error {
	return atomicLink(b.closedPath, b.linkPath)
}

func (b *fileBackend) Coffee() error {
	return atomicLink(b.coffeePath, b.linkPath)
}

func (b *fileBackend) Fire() error {
	return atomicLink(b.firePath, b.linkPath)
}
