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
var ConnectionSendStartRequest = PacketType(6)
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
		cbw, err := utils.WriteVarInt(w, p.ConnectionID)
		bw += cbw
		if err != nil {
			return int64(bw), err
		}
		if p.Type == ConnectionData {
			cbw, err := utils.WriteVarInt(w, len(p.Data))
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

func (p *Packet) ReadFrom(r utils.BufReader) (err error) {
	br, err := r.ReadByte()
	if err != nil {
		return err
	}
	p.Type = PacketType(br)
	if p.Type == Ingester || p.Type == Sender {
		return errors.New("invalid packet type")
	}
	if p.Type != ConnectionStartRequest && p.Type != ConnectionReset && p.Type != ConnectionKeepAlive {
		p.ConnectionID, err = utils.ReadVarInt(r)
		if err != nil {
			return err
		}
		if p.Type == ConnectionData {
			sz, err := utils.ReadVarInt(r)
			if err != nil {
				return err
			}
			p.Data = make([]byte, sz)
			_, err = io.ReadFull(r, p.Data)
			if err != nil {
				return err
			}
		}
	}
	return err
}
