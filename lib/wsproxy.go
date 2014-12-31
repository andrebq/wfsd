package lib

import (
	"code.google.com/p/go.net/websocket"
	"code.google.com/p/goplan9/plan9"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	KiB = 1024
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

	if len(conn.Request().URL.Query().Get("keepalive")) > 0 {
		isecs, err := strconv.ParseInt(conn.Request().URL.Query().Get("keepalive"), 32, 10)
		var keepalive time.Duration
		if err != nil {
			keepalive = 5 * time.Second
		} else {
			keepalive = time.Duration(isecs) * time.Second
		}
		pconn.(*net.TCPConn).SetKeepAlive(true)
		pconn.(*net.TCPConn).SetKeepAlivePeriod(keepalive)
	}

	conn.PayloadType = websocket.BinaryFrame
	// TODO: Send an Rerror instead of simply closing it
	println("waiting for Tversion")
	fc, err := plan9.ReadFcall(conn)
	if err != nil {
		pconn.Close()
		conn.Close()
		return
	}
	println("sending tversion to remote server")
	err = plan9.WriteFcall(pconn, fc)
	if err != nil {
		pconn.Close()
		conn.Close()
		return
	}
	// reading the response
	println("reading rversion response")
	fc, err = plan9.ReadFcall(pconn)
	if err != nil {
		pconn.Close()
		conn.Close()
		return
	}
	println("sending rversion to client")
	err = plan9.WriteFcall(conn, fc)
	if err != nil {
		pconn.Close()
		conn.Close()
		return
	}
	errch := make(chan error, 2)

	proxy := func(from, to net.Conn, errch chan error) {
		for {
			fc, err := plan9.ReadFcall(from)
			if err != nil {
				errch <- err
				return
			}
			println("[", from.RemoteAddr().String(), "] >>> ", fc.String())
			err = plan9.WriteFcall(to, fc)
			if err != nil {
				errch <- err
				return
			}
		}
	}

	go proxy(conn, pconn, errch)
	go proxy(pconn, conn, errch)

	<-errch
	<-errch
}

func NewWSProxy() http.Handler {
	return websocket.Handler(wsProxy)
}
