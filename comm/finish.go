package comm

import (
	"errors"
	"io"
)

var Finish = PacketType(128)

type FinishPacket struct{}

func (p *FinishPacket) WriteTo(w io.Writer) (n int64, err error) {
	bw, err := w.Write([]byte{byte(Finish)})
	return int64(bw), err
}

func (p *FinishPacket) ReadFrom(r io.Reader) (n int64, err error) {
	tbuff := make([]byte, 1)
	br, err := io.ReadFull(r, tbuff)
	if err != nil {
		return int64(br), err
	}
	if tbuff[0] != byte(Finish) {
		return int64(br), errors.New("invalid packet type")
	}
	return int64(br), nil
}
