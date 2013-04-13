package lib

import (
	"code.google.com/p/go.net/websocket"
)

func wsProxy(remote string, log LogFn) websocket.Handler {
	proxyFn := func(ws *websocket.Conn) {
		// open 9p session
		// at this moment, just give an salute message
		log("Opening new connection to: %v", remote)
		cli, err := NewWfsClient(remote)
		if err != nil {
			log("Error opening connection to: %v. Cause: %v", remote, err)
			return
		}
		defer cli.Close()
		for {
			msg := &WfsMessage{}
			err = msg.ReadFrom(ws)
			if err != nil {
				log("error: %v", err)
				return
			}
			msg = cli.Process(msg)
			err = msg.WriteTo(ws)
			if err != nil {
				log("Error: %v", err)
				return
			}
		}
	}
	return websocket.Handler(proxyFn)
}

func register9Proxy(prefix, remote string, mux Mux, log LogFn) {
	if len(remote) == 0 {
		log("Invalid config - Cannot create a 9Proxy at %v without a destination", prefix, remote)
		return
	}
	log("9Proxy - url: %v to %v", prefix, remote)
	mux.Handle(prefix, wsProxy(remote, log))
}
