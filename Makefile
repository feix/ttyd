.PHONY: build server ssh

GOFLAG=-ldflags "-s -w " -v

build: bin

bin: server ssh

server:
	go build $(GOFLAG) -o bin/server cmd/server.go

ssh:
	go build $(GOFLAG) -o bin/ssh cmd/ssh.go

clean:
	rm -rf bin/
