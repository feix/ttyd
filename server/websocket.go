package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/feix/ttyd/proxy"
	"github.com/feix/ttyd/service"
	"github.com/gorilla/websocket"
)

var upgrader *websocket.Upgrader

func init() {
	upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
}

func jsonError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "err": err.Error()})
}

func handleError(ws *websocket.Conn, username string, code int, err error) {
	log.Printf("websocket connection failed, %s, %d:%v", username, code, err)

	dt := time.Now().Add(time.Second)
	if err := ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, err.Error()), dt); err != nil {
		log.Printf("websocket writes control message failed: %v", err)
	}
}

func queryInt(r *http.Request, key string, defaultValue int) int {
	if valStr := r.URL.Query().Get(key); len(valStr) > 0 {
		if value, err := strconv.Atoi(valStr); err == nil {
			return value
		}
	}
	return defaultValue
}

// 准备 sshConfig
func querySSHConfig(r *http.Request) *service.SSHConfig {
	port := queryInt(r, "port", 22)
	loginStr := r.URL.Query().Get("login")
	if sshConfig := service.SSHConfigParser(loginStr, port); sshConfig != nil {
		return sshConfig
	}

	user := r.URL.Query().Get("user")
	host := r.URL.Query().Get("host")
	return &service.SSHConfig{
		User: user,
		Host: host,
		Port: port,
	}
}

func IsAuthorized(ctx context.Context, sshConfig service.SSHConfig) error {
	return nil
}

func WebSocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		jsonError(w, err)
		return
	}
	defer wsConn.Close()

	sshConfig := querySSHConfig(r)
	if err := IsAuthorized(ctx, *sshConfig); err != nil {
		handleError(wsConn, sshConfig.User, 4003, err)
		return
	}

	winSizeCh := make(chan *service.WinSize)

	wsSSH := service.NewWebSocketSSH(ctx, wsConn, winSizeCh)
	defer wsSSH.Close()

	if err := proxy.SSH(wsSSH.Context(), wsSSH, *sshConfig, winSizeCh); err != nil {
		handleError(wsConn, sshConfig.User, websocket.CloseInternalServerErr, err)
	}
}
