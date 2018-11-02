package main


import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"github.com/getsentry/raven-go"
	"github.com/WhiteHatCP/seclab-listener/backend"
	"github.com/WhiteHatCP/seclab-listener/server"
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
	if len(os.Args)%5 != 2 {
		fmt.Fprintf(os.Stderr, "usage: seclab key [dest open closed coffee fire [..]]\n")
		return
	}
	keypath := os.Args[1]
	s := server.New(keypath, maxPacketAge)

	for i := 2; i+2 < len(os.Args); i += 5 {
		dest := os.Args[i]
		openfile := os.Args[i+1]
		closedfile := os.Args[i+2]
		coffeefile := os.Args[i+3]
		firefile := os.Args[i+4]
		s.AddBackend(backend.New(dest, openfile, closedfile, coffeefile, firefile))
	}

	syscall.Umask(0007)

	socket := "seclab.sock"
	ln, err := openSock(socket)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
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
