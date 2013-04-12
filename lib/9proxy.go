package lib

import (
	"encoding/binary"
	"errors"
	"io"
)

type Signal struct{}

type NinePSession struct {
	client io.ReadWriteCloser
	remote io.ReadWriteCloser
	Done   <-chan Signal
	done   chan Signal
	stop   bool
}

func CreateSession(client, remote io.ReadWriteCloser) *NinePSession {
	ret := &NinePSession{client, remote, nil, make(chan Signal), false}
	// write only signal
	ret.Done = ret.done
	go ret.clientToRemote()
	go ret.remoteToClient()
	return ret
}

func (s *NinePSession) Start() {
	<-s.done
	s.stop = true
}

func proxyMessage(in, out io.ReadWriteCloser, buf []byte) ([]byte, error) {
	sz := int32(0)
	err := binary.Read(in, binary.BigEndian, &sz)
	if err != nil {
		return buf, err
	}
	err = binary.Write(out, binary.BigEndian, sz)
	if err != nil {
		return buf, err
	}
	if cap(buf) < int(sz) {
		buf = make([]byte, int(sz))
	} else {
		buf = buf[:int(sz)]
	}
	n, err := in.Read(buf)
	if n != int(sz) {
		return buf, errors.New("invalid size")
	}
	n, err = out.Write(buf)
	if n != int(sz) {
		return buf, errors.New("invalid size")
	}
	if err != nil {
		return buf, err
	}
	return buf, nil
}

func (s *NinePSession) clientToRemote() {
	// need better error handling instead of simply breaking
	// everything
	// but at this point, this should serve as a proof of concept
	buf := make([]byte, 1024)
	for !s.stop {
		var err error
		buf, err = proxyMessage(s.client, s.remote, buf)
		if err != nil {
			return
		}
	}
}

func (s *NinePSession) remoteToClient() {
	// need better error handling instead of simply breaking
	// everything
	// but at this point, this should serve as a proof of concept
	buf := make([]byte, 1024)
	for !s.stop {
		var err error
		buf, err = proxyMessage(s.remote, s.client, buf)
		if err != nil {
			return
		}
	}
}
