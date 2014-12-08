package lib

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
)

func wsProxy(conn *websocket.Conn) {
	ep := conn.Request().URL.Query().Get("endpoint")
	if len(ep) == 0 {
		io.WriteString(conn, "ERR no endpoint")
		conn.Close()
		return
	}
	epurl, err := url.Parse(ep)
	if err != nil {
		io.WriteString(conn, "ERR invalid endpoint")
		conn.Close()
		return
	}
	pconn, err := net.Dial(epurl.Scheme, epurl.Host)
	if err != nil {
		fmt.Fprintf(conn, "ERR unable to connect: %v", err)
		conn.Close()
		return
	}
	defer pconn.Close()
	_, err = io.WriteString(conn, "OK ")
	if err != nil {
		conn.Close()
		return
	}
	defer conn.Close()

	// everything fine until here, let's start the copy
	errch := make(chan error, 0)
	go func() {
		_, err := io.Copy(conn, pconn)
		errch <- err
	}()

	go func() {
		_, err := io.Copy(pconn, conn)
		errch <- err
	}()
	<-errch
	<-errch
}

func NewWSProxy() http.Handler {
	return websocket.Handler(wsProxy)
}
