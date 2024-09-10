package comm

import (
	"errors"
	"github.com/1f349/melon-backup/utils"
	"io"
)

var Sender = PacketType(0)

type SenderPacket struct {
	Mode                   int
	Services               *ServiceList
	RequestReboot          bool
	RequestServiceStop     bool
	RequestServiceRestart  bool
	RequestServiceStartNew bool
}

func (p *SenderPacket) WriteTo(w io.Writer) (n int64, err error) {
	bw, err := utils.WriteCompressedInt(p.Mode, w)
	if err != nil {
		return int64(bw), err
	}
	cbw, err := w.Write([]byte{byte(Sender)})
	bw += cbw
	if err != nil {
		return int64(bw), err
	}
	fbuff := make([]byte, 1)
	if p.Services != nil && len(p.Services.List) > 0 {
		fbuff[0] = 1
	}
	if p.RequestReboot {
		fbuff[0] += 2
	}
	if p.RequestServiceStop {
		fbuff[0] += 4
	}
	if p.RequestServiceRestart {
		fbuff[0] += 8
	}
	if p.RequestServiceStartNew {
		fbuff[0] += 16
	}
	cbw, err = w.Write(fbuff)
	bw += cbw
	if err != nil {
		return int64(bw), err
	}
	if p.Services != nil && len(p.Services.List) > 0 {
		cbw, err := p.Services.WriteTo(w)
		bw += int(cbw)
		if err != nil {
			return int64(bw), err
		}
	}
	return int64(bw), nil
}

func (p *SenderPacket) ReadFrom(r io.Reader) (n int64, err error) {
	var br int
	br, err, p.Mode = utils.ReadCompressedInt(r)
	if err != nil {
		return int64(br), err
	}
	tbuff := make([]byte, 1)
	cbr, err := io.ReadFull(r, tbuff)
	br += cbr
	if err != nil {
		return int64(br), err
	}
	if tbuff[0] != byte(Sender) {
		return int64(br), errors.New("invalid packet type")
	}
	fbuff := make([]byte, 1)
	cbr, err = io.ReadFull(r, tbuff)
	br += cbr
	if err != nil {
		return int64(br), err
	}
	p.RequestReboot = (fbuff[0] & 2) != 0
	p.RequestServiceStop = (fbuff[0] & 4) != 0
	p.RequestServiceRestart = (fbuff[0] & 8) != 0
	p.RequestServiceStartNew = (fbuff[0] & 16) != 0
	if fbuff[0]&1 != 0 {
		if p.Services == nil {
			p.Services = &ServiceList{}
		}
		crb, err := p.Services.ReadFrom(r)
		br += int(crb)
		if err != nil {
			return int64(br), err
		}
	}
	return int64(br), nil
}
