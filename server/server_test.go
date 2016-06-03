package server_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"testing"
	"time"
	"github.com/WhiteHatCP/seclab-listener/server"
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
	_, err := getTestInstance().ParseMessage(msg)
	if err == nil || err.Error() != "Incorrect HMAC signature" {
		t.Error("Expected Incorrect HMAC signature, got", err)
	}
}

func TestExpired(t *testing.T) {
	payload := make([]byte, 9)
	mac := hmac.New(sha256.New, []byte("dismykey"))
	mac.Write(payload)
	_, err := getTestInstance().ParseMessage(mac.Sum(payload))
	if err == nil || err.Error() != "Request expired" {
		t.Error("Expected Request expired, got", err)
	}
}

func getMessage(status uint8) []byte {
	payload := make([]byte, 9)
	payload[0] = status
	now64 := time.Now().Unix()
	binary.BigEndian.PutUint64(payload[1:9], uint64(now64))
	mac := hmac.New(sha256.New, []byte("dismykey"))
	mac.Write(payload)
	return mac.Sum(payload)
}

func TestOpenStatus(t *testing.T) {
	status, err := getTestInstance().ParseMessage(getMessage(0xff))
	if err != nil {
		t.Error(err)
	} else if status != 0xff {
		t.Error(status, "!= 0xff")
	}
}

func TestCloseStatus(t *testing.T) {
	status, err := getTestInstance().ParseMessage(getMessage(0x00))
	if err != nil {
		t.Error(err)
	} else if status != 0x00 {
		t.Error(status, "!= 0x00")
	}
}
