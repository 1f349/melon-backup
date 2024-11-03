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
	cbw, err := utils.WriteVarInt(w, p.Mode)
	bw += cbw
	return int64(bw), err
}

func (p *IngesterPacket) ReadFrom(r io.ByteReader) (err error) {
	br, err := r.ReadByte()
	if err != nil {
		return err
	}
	if br != byte(Ingester) {
		return errors.New("invalid packet type")
	}
	p.Mode, err = utils.ReadVarInt(r)
	return err
}
