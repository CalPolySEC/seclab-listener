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
	ln, err := net.Listen("unix", socket)
	syscall.Chmod(socket, 0770)
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
