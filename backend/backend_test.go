package backend_test

import (
	"bytes"
	"github.com/WhiteHatCP/seclab-listener/backend"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestBadPaths(t *testing.T) {
	b := backend.New("fakelink", "fakeopen", "fakeclose")
	if err := b.Open(); err == nil {
		t.Error("Expected LineError")
	}
	if err := b.Close(); err == nil {
		t.Error("Expected LineError")
	}
}

func TestLink(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("TempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	link := filepath.Join(tempDir, "status.txt")
	open := filepath.Join(tempDir, "open.txt")
	closed := filepath.Join(tempDir, "close.txt")
	if err := ioutil.WriteFile(open, []byte("Lab Open"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := ioutil.WriteFile(closed, []byte("Lab Closed"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Try to open
	b := backend.New(link, open, closed)
	if err := b.Open(); err != nil {
		t.Error(err)
	}
	data, err := ioutil.ReadFile(link)
	if err != nil {
		t.Error(err)
	} else if bytes.Compare(data, []byte("Lab Open")) != 0 {
		t.Error("Expected Lap Open, got", string(data))
	}

	// Try to close
	if err := b.Close(); err != nil {
		t.Error(err)
	}
	data, err = ioutil.ReadFile(link)
	if err != nil {
		t.Error(err)
	} else if bytes.Compare(data, []byte("Lab Closed")) != 0 {
		t.Error("Expected Lap Closed, got", string(data))
	}
}