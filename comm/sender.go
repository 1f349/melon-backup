package comm

import (
	"errors"
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/utils"
	"github.com/charmbracelet/log"
	"io"
	"strconv"
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
	bw, err := w.Write([]byte{byte(Sender)})
	if err != nil {
		return int64(bw), err
	}
	if conf.Debug {
		log.Error("pk_s_wt : a")
	}
	cbw, err := utils.WriteCompressedInt(p.Mode, w)
	bw += cbw
	if err != nil {
		return int64(bw), err
	}
	if conf.Debug {
		log.Error("pk_s_wt : b")
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
	if conf.Debug {
		log.Error("pk_s_wt : c")
	}
	if p.Services != nil && len(p.Services.List) > 0 {
		cbw, err := p.Services.WriteTo(w)
		bw += int(cbw)
		if err != nil {
			return int64(bw), err
		}
		if conf.Debug {
			log.Error("pk_s_wt : d")
		}
	}
	return int64(bw), nil
}

func (p *SenderPacket) ReadFrom(r io.Reader) (n int64, err error) {
	var cbr int
	tbuff := make([]byte, 1)
	br, err := io.ReadFull(r, tbuff)
	if err != nil {
		return int64(br), err
	}
	if tbuff[0] != byte(Sender) {
		return int64(br), errors.New("invalid packet type")
	}
	if conf.Debug {
		log.Error("pk_s_rf : a")
	}
	cbr, err, p.Mode = utils.ReadCompressedInt(r)
	if conf.Debug {
		log.Error("pk_s_rf : rci : " + strconv.Itoa(cbr))
	}
	br += cbr
	if err != nil {
		return int64(br), err
	}
	if conf.Debug {
		log.Error("pk_s_rf : b")
	}
	fbuff := make([]byte, 1)
	cbr, err = io.ReadFull(r, fbuff)
	br += cbr
	if err != nil {
		return int64(br), err
	}
	if conf.Debug {
		log.Error("pk_s_rf : c")
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
		if conf.Debug {
			log.Error("pk_s_rf : d")
		}
	}
	return int64(br), nil
}
