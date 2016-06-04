package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/WhiteHatCP/seclab-listener/backend"
	"io"
	"log"
	"net"
	"time"
)

const (
	ReqOpen     = 0xff
	ReqClose    = 0x00
	ReqKeygen   = 0xaa
	RespAllGood = 0xff
)

type Server interface {
	CheckMessage([]byte) error
	DispatchRequest(byte) ([]byte, error)
	Serve(net.Listener)
}

type server struct {
	key     []byte
	maxAge  int
	backend backend.Backend
}

func New(key []byte, maxAge int, backend backend.Backend) Server {
	return &server{
		key:     key,
		maxAge:  maxAge,
		backend: backend,
	}
}

func checkHash(key []byte, payload []byte, hash []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	expected := mac.Sum(nil)
	return hmac.Equal(hash, expected)
}

// Read the status byte, validate HMAC and timestamp
func (s *server) CheckMessage(data []byte) error {
	if !checkHash(s.key, data[:9], data[9:]) {
		return errors.New("Incorrect HMAC signature")
	}
	ts := int64(binary.BigEndian.Uint64(data[1:9]))
	if time.Now().Unix()-ts > int64(s.maxAge) {
		return errors.New("Request expired")
	}
	return nil
}

func (s *server) DispatchRequest(status byte) ([]byte, error) {
	if status == ReqOpen {
		log.Print("Received request: open")
		return []byte{RespAllGood}, s.backend.Open()
	} else if status == ReqClose {
		log.Print("Received request: close")
		return []byte{RespAllGood}, s.backend.Close()
	} else if status == ReqKeygen {
		return nil, fmt.Errorf("Keygen not implemented")
	} else {
		return nil, fmt.Errorf("Unrecognized status byte: 0x%02x", status)
	}
}

func (s *server) handleConnection(conn net.Conn) {
	defer conn.Close()
	data := make([]byte, 41)
	for {
		if _, err := io.ReadFull(conn, data); err != nil {
			log.Print(err)
			return
		}
		err := s.CheckMessage(data)
		if err != nil {
			log.Print(err)
			return
		}
		resp, err := s.DispatchRequest(data[0])
		if err != nil {
			log.Print(err)
			return
		}
		conn.Write(resp)
	}
}

func (s *server) Serve(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.handleConnection(conn)
	}
}
