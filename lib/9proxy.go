package lib

import (
	"code.google.com/p/go9p/p"
	"code.google.com/p/go9p/p/clnt"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"unicode/utf8"
)

const (
	WfsOpen   = 0
	WfsRead   = 1
	WfsWrite  = 2
	WfsClose  = 3
	WfsCreate = 4

	WfsModeRead      = p.OREAD
	WfsModeWrite     = p.OWRITE
	WfsModeReadWrite = p.ORDWR
	WfsModeTruncate  = p.OTRUNC

	WfsModeDir = p.DMDIR

	WfsPermWrite = p.DMWRITE
	WfsPermRead  = p.DMREAD
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
	case WfsRead:
		return w.ReadFile(msg)
	case WfsClose:
		return w.CloseFile(msg)
	case WfsCreate:
		return w.CreateFile(msg)
		/*case WfsWrite: return w.WriteFile(msg)
		 */
	}
	msg.Error = "Invalid message type"
	return msg
}

func (w *WfsClient) CreateFile(msg *WfsMessage) *WfsMessage {
	return msg
}

func (w *WfsClient) CloseFile(msg *WfsMessage) *WfsMessage {
	if _, has := w.UseFid(msg.Fid); has {
		w.CloseFid(msg.Fid)
		return msg
	}
	msg.Error = "fid not found"
	return msg
}

func (w *WfsClient) ReadFile(msg *WfsMessage) *WfsMessage {
	if file, has := w.UseFid(msg.Fid); has {
		buf, err := ioutil.ReadAll(file)
		if err != nil {
			msg.Error = err.Error()
		} else {
			if utf8.Valid(buf) {
				msg.Data = string(buf)
			} else {
				// use base64
				msg.Error = "Not possible at this moment"
			}
		}
	} else {
		msg.Error = "Fid isn't open"
	}
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
		w.BindFid(msg.Fid, file)
	default:
		msg.Error = "Invalid mode"
	}
	return msg
}

func (w *WfsClient) BindFid(fid int32, file *clnt.File) {
	w.files[fid] = file
}

func (w *WfsClient) UseFid(fid int32) (*clnt.File, bool) {
	file, has := w.files[fid]
	return file, has
}

func (w *WfsClient) CloseFid(fid int32) {
	if file, has := w.UseFid(fid); has {
		file.Close()
		delete(w.files, fid)
	}
}
