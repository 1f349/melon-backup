package utils

import (
	"errors"
	"io"
)

func WriteCompressedInt(i int, writer io.Writer) (n int, err error) {
	if i < 0 {
		return 0, errors.New("negative int")
	} else if i == 0 {
		bw, err := writer.Write([]byte{0})
		return bw, err
	}
	bw := 0
	currentI := i
	var nextI int
	var currentMask int
	for currentI > 0 {
		nextI = currentI >> 7
		currentMask = nextI << 7
		var bt = byte(currentI - currentMask)
		if nextI > 0 {
			bt |= 128
		}
		cbw, err := writer.Write([]byte{bt})
		bw += cbw
		if err != nil {
			return bw, err
		}
		currentI = nextI
	}
	return bw, nil
}

func ReadCompressedInt(r io.Reader) (n int, err error, val int) {
	cBuff := make([]byte, 1)
	br, err := io.ReadFull(r, cBuff)
	if err != nil {
		return br, err, 0
	} else if cBuff[0] == 0 {
		return br, nil, 0
	}
	moreBytes := true
	valToRet := 0
	cBitSize := 0
	cbr := 0
	for moreBytes {
		moreBytes = (cBuff[0] & 128) != 0
		cBuff[0] &^= 128
		valToRet += int(uint64(cBuff[0]) << cBitSize)
		cBitSize += 7
		cbr, err = io.ReadFull(r, cBuff)
		br += cbr
		if err != nil {
			return br, err, valToRet
		}
	}
	return br, nil, valToRet
}
