package server_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"github.com/WhiteHatCP/seclab-listener/server"
	"testing"
	"time"
)

type nullBackend struct{}

func (b *nullBackend) Open() error {
	return nil
}
func (b *nullBackend) Close() error {
	return nil
}

type errorBackend struct{}

func (b *errorBackend) Open() error {
	return errors.New("open error")
}
func (b *errorBackend) Close() error {
	return errors.New("close error")
}

func getTestInstance() server.Server {
	return server.New([]byte("dismykey"), 10, &nullBackend{})
}

func TestBadSignature(t *testing.T) {
	msg := make([]byte, 41)
	err := getTestInstance().CheckMessage(msg)
	if err == nil || err.Error() != "Incorrect HMAC signature" {
		t.Error("Expected Incorrect HMAC signature, got", err)
	}
}

func TestExpired(t *testing.T) {
	payload := make([]byte, 9)
	mac := hmac.New(sha256.New, []byte("dismykey"))
	mac.Write(payload)
	err := getTestInstance().CheckMessage(mac.Sum(payload))
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
	if err := getTestInstance().CheckMessage(message); err != nil {
		t.Error(err)
	}
}

func TestDispatchUnknown(t *testing.T) {
	_, err := getTestInstance().DispatchRequest(0x69)
	if err == nil || err.Error() != "Unrecognized status byte: 0x69" {
		t.Error("Expected Unrecognized status byte: 0x69, got", err)
	}
}

func TestDispatchOpenError(t *testing.T) {
	s := server.New(nil, 10, &errorBackend{})
	_, err := s.DispatchRequest(0xff)
	if err == nil || err.Error() != "open error" {
		t.Error("Expected open error, got", err)
	}
}

func TestDispatchCloseError(t *testing.T) {
	s := server.New(nil, 10, &errorBackend{})
	_, err := s.DispatchRequest(0x00)
	if err == nil || err.Error() != "close error" {
		t.Error("Expected close error, got", err)
	}
}

func TestDispatchOpenGood(t *testing.T) {
	resp, err := getTestInstance().DispatchRequest(0xff)
	if err != nil {
		t.Error(err)
	} else if resp != 0xff {
		t.Errorf("Expected 0xff, got 0x%x", resp)
	}
}

func TestDispatchCloseGood(t *testing.T) {
	resp, err := getTestInstance().DispatchRequest(0x00)
	if err != nil {
		t.Error(err)
	} else if resp != 0xff {
		t.Errorf("Expected 0xff, got 0x%x", resp)
	}
}
