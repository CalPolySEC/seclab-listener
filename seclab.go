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

func main() {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "usage: seclab key dest open closed\n")
		return
	}
	keypath := os.Args[1]
	dest := os.Args[2]
	openfile := os.Args[3]
	closedfile := os.Args[4]

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

	backend := backend.New(dest, openfile, closedfile)
	s := server.New(keypath, 10, backend)
	s.Serve(ln)
}
