package comm

import (
	"errors"
	"github.com/1f349/melon-backup/utils"
	"io"
)

var Ingester = PacketType(255)

type IngesterPacket struct {
	Mode int
}

func (p *IngesterPacket) WriteTo(w io.Writer) (n int64, err error) {
	bw, err := w.Write([]byte{byte(Ingester)})
	if err != nil {
		return int64(bw), err
	}
	cbw, err := utils.WriteIntAsBytes(p.Mode, w)
	bw += cbw
	return int64(bw), err
}

func (p *IngesterPacket) ReadFrom(r io.Reader) (n int64, err error) {
	tbuff := make([]byte, 1)
	br, err := io.ReadFull(r, tbuff)
	if err != nil {
		return int64(br), err
	}
	if tbuff[0] != byte(Ingester) {
		return int64(br), errors.New("invalid packet type")
	}
	var cbr int
	cbr, err, p.Mode = utils.ReadIntFromBytes(r)
	br += cbr
	return int64(br), err
}
