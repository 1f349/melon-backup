package comm

import (
	"errors"
	"github.com/1f349/melon-backup/utils"
	"io"
)

type PacketType byte

var ConnectionStartRequest = PacketType(1)
var ConnectionStarted = PacketType(2)
var ConnectionReset = PacketType(3)
var ConnectionClosed = PacketType(4)
var ConnectionData = PacketType(5)
var ConnectionKeepAlive = PacketType(254)

type Packet struct {
	Type         PacketType
	ConnectionID int
	Data         []byte
}

func (p *Packet) WriteTo(w io.Writer) (n int64, err error) {
	bw, err := w.Write([]byte{byte(p.Type)})
	if err != nil {
		return int64(bw), err
	}
	if p.Type == Ingester || p.Type == Sender {
		return int64(bw), errors.New("invalid packet type")
	}
	if p.Type != ConnectionStartRequest && p.Type != ConnectionReset && p.Type != ConnectionKeepAlive {
		cbw, err := utils.WriteCompressedInt(p.ConnectionID, w)
		bw += cbw
		if err != nil {
			return int64(bw), err
		}
		if p.Type == ConnectionData {
			cbw, err := utils.WriteCompressedInt(len(p.Data), w)
			bw += cbw
			if err != nil {
				return int64(bw), err
			}
			cbw, err = w.Write(p.Data)
			bw += cbw
			if err != nil {
				return int64(bw), err
			}
		}
	}
	return int64(bw), nil
}

func (p *Packet) ReadFrom(r io.Reader) (n int64, err error) {
	tbuff := make([]byte, 1)
	br, err := io.ReadFull(r, tbuff)
	if err != nil {
		return int64(br), err
	}
	p.Type = PacketType(tbuff[0])
	if p.Type == Ingester || p.Type == Sender {
		return int64(br), errors.New("invalid packet type")
	}
	if p.Type != ConnectionStartRequest && p.Type != ConnectionReset && p.Type != ConnectionKeepAlive {
		var cbr int
		cbr, err, p.ConnectionID = utils.ReadCompressedInt(r)
		br += cbr
		if err != nil {
			return int64(br), err
		}
		if p.Type == ConnectionData {
			cbr, err, sz := utils.ReadCompressedInt(r)
			br += cbr
			if err != nil {
				return int64(br), err
			}
			p.Data = make([]byte, sz)
			cbr, err = io.ReadFull(r, p.Data)
			br += cbr
			if err != nil {
				return int64(br), err
			}
		}
	}
	return int64(br), err
}
