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
	OpenStatus  = 0xff
	CloseStatus = 0x00
)

type Server interface {
	ParseMessage([]byte) (uint8, error)
	Serve(string)
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

func (s *server) ParseMessage(data []byte) (uint8, error) {
	if !checkHash(s.key, data[:9], data[9:]) {
		return 0, errors.New("Incorrect HMAC signature")
	}

	status, ts := data[0], int64(binary.BigEndian.Uint64(data[1:9]))
	age := time.Now().Unix() - ts
	if age > int64(s.maxAge) {
		return 0, errors.New("Request expired")
	}

	return status, nil
}

func (s *server) handleConnection(conn net.Conn) {
	defer conn.Close()

	data := make([]byte, 41)
	if _, err := io.ReadFull(conn, data); err != nil {
		log.Print(err)
		return
	}

	status, err := s.ParseMessage(data)
	if err != nil {
		log.Print(err)
		return
	}

	if status == OpenStatus {
		log.Print("Received request: open")
		err = s.backend.Open()
	} else if status == CloseStatus {
		log.Print("Received request: close")
		err = s.backend.Close()
	} else {
		err = fmt.Errorf("Unrecognized status byte: 0x%02x", status)
	}

	if err != nil {
		log.Print(err)
		return
	}

	// Success, send response
	if _, err := conn.Write([]byte("\xff")); err != nil {
		log.Print(err)
		return
	}
}

func (s *server) Serve(port string) {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.handleConnection(conn)
	}
}
