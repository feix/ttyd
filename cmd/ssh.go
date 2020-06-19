package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/feix/ttyd/proxy"
	"github.com/feix/ttyd/service"
	"golang.org/x/crypto/ssh/terminal"
)

func usage() {
	fmt.Printf(`Usage:
	%s loginString

Example:
		%s user@host
`, os.Args[0], os.Args[0])
}

func main() {
	ctx := context.Background()
	loginStr := os.Args[1]
	port := 2223
	if len(os.Args) > 2 {
		port, _ = strconv.Atoi(os.Args[2])
	}
	sshConfig := service.SSHConfigParser(loginStr, port)
	if sshConfig == nil {
		fmt.Printf("%s, incorrect format loginStr", loginStr)
		usage()
		return
	}

	winSizeCh := make(chan *service.WinSize)
	width, height, _ := terminal.GetSize(0)
	winSizeCh <- &service.WinSize{Width: width, Height: height}

	rw := struct {
		io.Writer
		io.Reader
	}{
		os.Stdout,
		os.Stdin,
	}

	err := proxy.SSH(ctx, rw, *sshConfig, winSizeCh)
	if err != nil {
		panic(err)
	}
}
