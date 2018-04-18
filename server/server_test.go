package server_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"github.com/WhiteHatCP/seclab-listener/server"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type Closer func()

type countingBackend struct {
	numOpen  int
	numClose int
}

func (b *countingBackend) Open() error {
	b.numOpen += 1
	return nil
}

func (b *countingBackend) Close() error {
	b.numClose += 1
	return nil
}

type errorBackend struct{}

func (b *errorBackend) Open() error {
	return errors.New("open error")
}
func (b *errorBackend) Close() error {
	return errors.New("close error")
}
func (b *errorBackend) Coffee() error {
	return errors.New("coffee error")
}

func getTestInstance() (server.Server, Closer) {
	tempDir, _ := ioutil.TempDir("", "")
	keypath := filepath.Join(tempDir, "key")
	ioutil.WriteFile(keypath, []byte("dismykey"), 0644)
	s := server.New(keypath, 10)
	s.AddBackend(&countingBackend{})
	return s, func() {
		os.RemoveAll(tempDir)
	}
}

func TestBadSignature(t *testing.T) {
	msg := make([]byte, 41)
	s, close := getTestInstance()
	defer close()
	err := s.CheckMessage(msg)
	if err == nil || err.Error() != "Incorrect HMAC signature" {
		t.Error("Expected Incorrect HMAC signature, got", err)
	}
}

func TestExpired(t *testing.T) {
	payload := make([]byte, 9)
	mac := hmac.New(sha256.New, []byte("dismykey"))
	mac.Write(payload)
	s, close := getTestInstance()
	defer close()
	err := s.CheckMessage(mac.Sum(payload))
	if err == nil || err.Error() != "Request expired" {
		t.Error("Expected Request expired, got", err)
	}
}

func TestGoodCheck(t *testing.T) {
	payload := make([]byte, 9)
	payload[0] = 0xff
	now64 := time.Now().Unix()
	binary.BigEndian.PutUint64(payload[1:9], uint64(now64))
	mac := hmac.New(sha256.New, []byte("dismykey"))
	mac.Write(payload)
	message := mac.Sum(payload)
	s, close := getTestInstance()
	defer close()
	if err := s.CheckMessage(message); err != nil {
		t.Error(err)
	}
}

func TestDispatchUnknown(t *testing.T) {
	s, close := getTestInstance()
	defer close()
	_, err := s.DispatchRequest(0x69)
	if err == nil || err.Error() != "Unrecognized status byte: 0x69" {
		t.Error("Expected Unrecognized status byte: 0x69, got", err)
	}
}

func TestDispatchOpenError(t *testing.T) {
	s := server.New("", 10)
	s.AddBackend(&errorBackend{})
	_, err := s.DispatchRequest(0xff)
	if err == nil || err.Error() != "open error" {
		t.Error("Expected open error, got", err)
	}
}

func TestDispatchCloseError(t *testing.T) {
	s := server.New("", 10)
	s.AddBackend(&errorBackend{})
	_, err := s.DispatchRequest(0x00)
	if err == nil || err.Error() != "close error" {
		t.Error("Expected close error, got", err)
	}
}

func TestDispatchOpenGood(t *testing.T) {
	s, close := getTestInstance()
	defer close()
	resp, err := s.DispatchRequest(0xff)
	if err != nil {
		t.Error(err)
	} else if resp[0] != 0xff {
		t.Errorf("Expected 0xff, got 0x%x", resp)
	}
}

func TestDispatchCloseGood(t *testing.T) {
	s, close := getTestInstance()
	defer close()
	resp, err := s.DispatchRequest(0x00)
	if err != nil {
		t.Error(err)
	} else if resp[0] != 0xff {
		t.Errorf("Expected 0xff, got 0x%x", resp)
	}
}

func TestKeyRotate(t *testing.T) {
	tempDir, _ := ioutil.TempDir("", "")
	keypath := filepath.Join(tempDir, "key")
	ioutil.WriteFile(keypath, []byte("dismykey"), 0644)
	s := server.New(keypath, 10)
	resp, err := s.KeyRotate()
	if err != nil {
		t.Error(err)
	}
	key, _ := ioutil.ReadFile(keypath)
	if bytes.Compare(resp[9:], key) != 0 {
		t.Error(resp, "!=", key)
	}
}

func TestMultipleBackends(t *testing.T) {
	b := &countingBackend{0, 0}
	s := server.New("", 10)
	s.AddBackend(b)
	s.AddBackend(b)

	if _, err := s.DispatchRequest(0x00); err != nil {
		t.Error(err)
	}
	if _, err := s.DispatchRequest(0xff); err != nil {
		t.Error(err)
	}

	if b.numOpen != 2 {
		t.Error("Expected Open to be called 2 times, was called ", b.numOpen)
	} else if b.numClose != 2 {
		t.Error("Expected Close to be called 2 times, was called ", b.numClose)
	}
}
