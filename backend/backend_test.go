package backend_test

import (
	"bytes"
	"github.com/WhiteHatCP/seclab-listener/backend"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBadPaths(t *testing.T) {
	b := backend.New("fakelink", "fakeopen", "fakeclose", "fakecoffee", "fakefire")
	if err := b.Open(); err == nil {
		t.Error("Expected LineError")
	}
	if err := b.Close(); err == nil {
		t.Error("Expected LineError")
	}
	if err := b.Coffee(); err == nil {
		t.Error("Expected LineError")
	}
	if err := b.Fire(); err == nil {
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
	coffee := filepath.Join(tempDir, "coffee.txt")
	fire := filepath.Join(tempDir, "fire.txt")
	if err := ioutil.WriteFile(open, []byte("Lab Open"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := ioutil.WriteFile(closed, []byte("Lab Closed"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := ioutil.WriteFile(coffee, []byte("Out for Coffee"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := ioutil.WriteFile(fire, []byte("Lab is Fire"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	os.Chtimes(open, time.Unix(0, 0), time.Unix(0, 0))

	// Try to open
	b := backend.New(link, open, closed, coffee)
	if err := b.Open(); err != nil {
		t.Error(err)
	}
	data, err := ioutil.ReadFile(link)
	if err != nil {
		t.Error(err)
	} else if bytes.Compare(data, []byte("Lab Open")) != 0 {
		t.Error("Expected Lap Open, got", string(data))
	}

	// Check that the mtime changed
	fi, err := os.Stat(link)
	if err != nil {
		t.Error(err)
	} else if fi.ModTime() == time.Unix(0, 0) {
		t.Error("Expected mtime to change")
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

	// Try to coffee
	if err := b.Coffee(); err != nil {
		t.Error(err)
	}
	data, err = ioutil.ReadFile(link)
	if err != nil {
		t.Error(err)
	} else if bytes.Compare(data, []byte("Out for Coffee")) != 0 {
		t.Error("Expected Out for Coffee, got", string(data))
	}

	// Try to fire
	if err := b.Fire(); err != nil {
		t.Error(err)
	}
	data, err = ioutil.ReadFile(link)
	if err != nil {
		t.Error(err)
	} else if bytes.Compare(data, []byte("Lab is Fire")) != 0 {
		t.Error("Expected Lab is Fire, got", string(data))
	}
}
