package main

import (
	"github.com/WhiteHatCP/seclab-listener/backend"
	"github.com/WhiteHatCP/seclab-listener/server"
	"os"
)

func main() {
	backend := backend.New("status.txt", "open.txt", "closed.txt")
	s := server.New([]byte(os.Args[1]), 10, backend)
	s.Serve(":5000")
}
