package main

import (
	"github.com/WhiteHatCP/seclab-listener/backend"
	"github.com/WhiteHatCP/seclab-listener/server"
	"log"
	"net"
	"os"
	"syscall"
)

func main() {
	syscall.Umask(0007)

	socket := "seclab.sock"
	ln, err := net.Listen("unix", socket)
	syscall.Chmod(socket, 0770)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	backend := backend.New("status.txt", "open.txt", "closed.txt")
	s := server.New([]byte(os.Args[1]), 10, backend)
	s.Serve(ln)
}
