package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/WhiteHatCP/seclab-listener/backend"
	"github.com/getsentry/raven-go"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

const (
	keyLength   = 32
	reqOpen     = 0xff
	reqClose    = 0x00
	reqCoffee	= 0xcc
	reqKeygen   = 0xaa
	respAllGood = 0xff
	respNewKey  = 0x55
)

var outLog = log.New(os.Stdout, "", log.LstdFlags)
var errLog = log.New(os.Stderr, "", log.LstdFlags)

// Server is the interface that handles the network protocol
type Server interface {
	AddBackend(backend.Backend)
	CheckMessage([]byte) error
	DispatchRequest(byte) ([]byte, error)
	KeyRotate() ([]byte, error)
	Serve(net.Listener)
}

type server struct {
	keypath  string
	maxAge   int
	backends []backend.Backend
}

// New creates a new instance of a Server
func New(keypath string, maxAge int) Server {
	return &server{
		keypath:  keypath,
		maxAge:   maxAge,
		backends: nil,
	}
}

func (s *server) AddBackend(b backend.Backend) {
	s.backends = append(s.backends, b)
}

func checkHash(key []byte, payload []byte, hash []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	expected := mac.Sum(nil)
	return hmac.Equal(hash, expected)
}

// Read the status byte, validate HMAC and timestamp
func (s *server) CheckMessage(data []byte) error {
	key, err := ioutil.ReadFile(s.keypath)
	if err != nil {
		return err
	}
	if !checkHash(key, data[:9], data[9:]) {
		return errors.New("Incorrect HMAC signature")
	}
	ts := int64(binary.BigEndian.Uint64(data[1:9]))
	if time.Now().Unix()-ts > int64(s.maxAge) {
		return errors.New("Request expired")
	}
	return nil
}

func (s *server) KeyRotate() ([]byte, error) {
	resp := make([]byte, 9+keyLength)
	key := resp[9:]
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(s.keypath, key, 0600); err != nil {
		return nil, err
	}
	resp[0] = respNewKey
	binary.BigEndian.PutUint64(resp[1:9], uint64(time.Now().Unix()))
	return resp, nil
}

func (s *server) open() error {
	for _, b := range s.backends {
		if err := b.Open(); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) close() error {
	for _, b := range s.backends {
		if err := b.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) coffee() error {
	for _, b :=  range s.backends {
		if err := b.Coffee(); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) DispatchRequest(status byte) ([]byte, error) {
	if status == reqOpen {
		outLog.Print("Received request: open")
		return []byte{respAllGood}, s.open()
	} else if status == reqClose {
		outLog.Print("Received request: close")
		return []byte{respAllGood}, s.close()
	} else if status == reqKeygen {
		return s.KeyRotate()
	}
	return nil, fmt.Errorf("Unrecognized status byte: 0x%02x", status)
}

func (s *server) readAndRespond(conn net.Conn) error {
	defer conn.Close()
	data := make([]byte, 9+keyLength)
	for {
		if _, err := io.ReadFull(conn, data); err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
		if err := s.CheckMessage(data); err != nil {
			return err
		}
		resp, err := s.DispatchRequest(data[0])
		if err != nil {
			return err
		}
		if _, err = conn.Write(resp); err != nil {
			return err
		}
	}
}

func (s *server) handleConnection(conn net.Conn) {
	if err := s.readAndRespond(conn); err != nil {
		raven.CaptureError(err, nil)
		errLog.Print(err)
	}
}

func (s *server) Serve(ln net.Listener) {
	outLog.Print("Seclab listener started")
	for {
		conn, err := ln.Accept()
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			errLog.Fatal(err)
		}
		go s.handleConnection(conn)
	}
}
