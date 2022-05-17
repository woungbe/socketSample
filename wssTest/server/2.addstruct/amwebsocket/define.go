package amwebsocket

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 81920

	liveTime = 60
)

var upgrader = websocket.Upgrader{ReadBufferSize: 81920, WriteBufferSize: 81920, EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

type ResultWebSocket struct {
	Command string
	Types   string
}

type WSPackageReq struct {
	PACKETVERSION string //패킷 버젼
	Command       string //커맨더  ( Login)
	LoginID       string //로그인 ID
}
