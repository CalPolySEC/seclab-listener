package main

import (
	"fmt"
	"github.com/WhiteHatCP/seclab-listener/backend"
	"github.com/WhiteHatCP/seclab-listener/server"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const (
	maxPacketAge = 10
)

func openSock(path string) (net.Listener, error) {
	if err := syscall.Unlink(path); err != nil && err != syscall.EEXIST {
		return nil, err
	}
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	if err := syscall.Chmod(path, 0770); err != nil {
		return nil, err
	}
	return ln, nil
}

func main() {
	if len(os.Args)%3 != 2 {
		fmt.Fprintf(os.Stderr, "usage: seclab key [dest open closed [..]]\n")
		return
	}
	keypath := os.Args[1]
	s := server.New(keypath, maxPacketAge)

	for i := 2; i+2 < len(os.Args); i += 3 {
		dest := os.Args[i]
		openfile := os.Args[i+1]
		closedfile := os.Args[i+2]
		s.AddBackend(backend.New(dest, openfile, closedfile))
	}

	syscall.Umask(0007)

	socket := "seclab.sock"
	ln, err := openSock(socket)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	// Make sure the socket closes
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for _ = range c {
			ln.Close()
		}
	}()

	s.Serve(ln)
}
