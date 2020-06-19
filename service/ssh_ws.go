package service

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	Input  = '0'
	Resize = '1'
	Ping   = '2'
)

type WinSize struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

type SSHAdapter struct {
	conn       *websocket.Conn
	readMutex  sync.Mutex
	writeMutex sync.Mutex
	reader     *bufio.Reader
	ctx        context.Context
	cancel     context.CancelFunc
	winSizeCh  chan<- *WinSize
}

func NewWebSocketSSH(ctx context.Context, conn *websocket.Conn, winSizeCh chan<- *WinSize) *SSHAdapter {
	c, cancel := context.WithCancel(ctx)

	return &SSHAdapter{
		conn:      conn,
		winSizeCh: winSizeCh,
		ctx:       c,
		cancel:    cancel,
	}
}

func (a *SSHAdapter) Context() context.Context {
	return a.ctx
}

func (a *SSHAdapter) Read(b []byte) (int, error) {
	// Read() can be called concurrently, and we mutate some internal state here
	a.readMutex.Lock()
	defer a.readMutex.Unlock()

	if a.reader == nil {
		_, reader, err := a.conn.NextReader()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				a.cancel()
			}

			return 0, err
		}

		a.reader = bufio.NewReader(reader)
	}

	msgType, err := a.reader.ReadByte()
	if err != nil {
		a.reader = nil

		// EOF for the current Websocket frame, more will probably come so..
		if err == io.EOF {
			// .. we must hide this from the caller since our semantics are a
			// stream of bytes across many frames
			err = nil
		}

		return 0, err
	}

	switch msgType {
	case Input:
		bs, e := ioutil.ReadAll(a.reader)
		return copy(b, bs), e
	case Resize:
		var winSize WinSize
		if err := json.NewDecoder(a.reader).Decode(&winSize); err == nil {
			a.winSizeCh <- &winSize
		}

		return 0, err
	case Ping:
		return 0, nil
	default:
		return 0, err
	}
}

func (a *SSHAdapter) Write(b []byte) (int, error) {
	a.writeMutex.Lock()
	defer a.writeMutex.Unlock()

	nextWriter, err := a.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}
	defer nextWriter.Close()

	return nextWriter.Write(b)
}

func (a *SSHAdapter) Close() error {
	return a.conn.Close()
}
