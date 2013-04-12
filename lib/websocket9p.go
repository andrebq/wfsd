package lib

import (
	"code.google.com/p/go.net/websocket"
	"io"
)

func Proxy(conn *websocket.Conn) {
	// open 9p session
	// at this moment, just give an salute message
	defer conn.Close()
	io.Copy(conn, conn)
}

func register9Proxy(prefix, remote string, mux Mux, log LogFn) {
	log("9Proxy - url: %v to %v", prefix, "put the target here")
	wsHandler := websocket.Handler(Proxy)
	mux.Handle(prefix, wsHandler)
}
