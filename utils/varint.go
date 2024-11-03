package utils

import (
	"encoding/binary"
	"io"
)

func WriteVarInt(w io.Writer, x int) (int, error) {
	var buf [binary.MaxVarintLen64]byte
	binary.PutVarint(buf[:], int64(x))
	return w.Write(buf[:])
}

func ReadVarInt(r io.ByteReader) (int, error) {
	n, err := binary.ReadVarint(r)
	return int(n), err
}
