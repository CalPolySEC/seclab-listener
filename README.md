Seclab Listener
===============

[![Build Status](https://api.travis-ci.org/WhiteHatCP/seclab-listener.svg?branch=master)](https://travis-ci.org/WhiteHatCP/seclab-listener)
[![Lab Status](https://thewhitehat.club/status.svg)](https://thewhitehat.club/)

## Building

To build the server program in the current directory:

```bash
mkdir -p $GOPATH/src/github.com/WhiteHatCP
git clone https://github.com/WhiteHatCP/seclab-listener.git $GOPATH/src/github.com/WhiteHatCP/seclab-listener
go build github.com/WhiteHatCP/seclab-listener
```

## Running
```bash
$ ./seclab-listener
usage: seclab key dest open closed
$ ./seclab-listener ~/private/key /var/www/status.svg /var/www/open.svg /var/www/closed.svg
```

You are now listening for new connections on `seclab.sock` in the current directory.
