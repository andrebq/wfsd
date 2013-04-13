package lib

import (
	"code.google.com/p/go9p/p"
	"code.google.com/p/go9p/p/clnt"
	"encoding/json"
	"io"
	"os"
)

const (
	WfsOpen  = 0
	WfsRead  = 1
	WfsWrite = 2
	WfsClose = 3

	WfsModeRead      = 0
	WfsModeWrite     = 1
	WfsModeReadWrite = 2
	WfsModeTruncate  = 16
)

type WfsMessage struct {
	Tag   int32
	Type  int32
	Mode  int32
	Fid   int32
	Error string
	Data  string
}

func (msg *WfsMessage) WriteTo(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(msg)
}

func (msg *WfsMessage) ReadFrom(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(msg)
}

type WfsClient struct {
	c     *clnt.Clnt
	files map[int32]*clnt.File
}

// Close the connection
func (w *WfsClient) Close() error {
	w.c.Unmount()
	w.files = nil
	w.c = nil
	return nil
}

// Open a new connection to the given remote host
func NewWfsClient(remote string) (*WfsClient, error) {
	cli := &WfsClient{nil, make(map[int32]*clnt.File)}
	user := p.OsUsers.Uid2User(os.Geteuid())
	var err error
	cli.c, err = clnt.Mount("tcp", remote, "", user)
	return cli, err
}

func (w *WfsClient) Process(msg *WfsMessage) *WfsMessage {
	switch msg.Type {
	case WfsOpen:
		return w.OpenFile(msg)
		/* case WfsRead: return w.ReadFile(msg)
		case WfsWrite: return w.WriteFile(msg)
		case WfsClose: return w.CloseFile(msg)
		*/
	}
	msg.Error = "Invalid message type"
	return msg
}

func (w *WfsClient) OpenFile(msg *WfsMessage) *WfsMessage {
	if _, has := w.files[msg.Fid]; has {
		msg.Error = "Fid already used"
		return msg
	}
	switch msg.Mode {
	case WfsModeTruncate, WfsModeReadWrite, WfsModeRead,
		WfsModeWrite:
		file, err := w.c.FOpen(msg.Data, uint8(msg.Mode))
		if err != nil {
			msg.Error = err.Error()
			return msg
		}
		w.UseFid(msg.Fid, file)
	default:
		msg.Error = "Invalid mode"
	}
	return msg
}

func (w *WfsClient) UseFid(fid int32, file *clnt.File) {
	w.files[fid] = file
}
