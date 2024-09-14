package utils

import (
	"errors"
	"io"
	"math"
)

var overflowError = errors.New("overflow")

func WriteIntAsBytes(i int, writer io.Writer) (n int, err error) {
	if i < 0 {
		return 0, errors.New("negative int")
	} else if i == 0 {
		bw, err := writer.Write([]byte{0})
		return bw, err
	}
	bw := 0
	currentI := i
	for currentI > 0 {
		var bt = byte(currentI & 127)
		currentI = currentI >> 7
		if currentI > 0 {
			bt |= 128
		}
		cbw, err := writer.Write([]byte{bt})
		bw += cbw
		if err != nil {
			return bw, err
		}
	}
	return bw, nil
}

func ReadIntFromBytes(r io.Reader) (n int, err error, val int) {
	cBuff := make([]byte, 1)
	valToRet := 0
	cBitSize := 0
	cbr := 0
	for valToRet < math.MaxInt {
		br, err := io.ReadFull(r, cBuff)
		cbr += br
		if err != nil {
			return cbr, err, valToRet
		}
		if cBuff[0] < 128 {
			if math.MaxInt-valToRet < int(cBuff[0]) {
				return cbr, overflowError, math.MaxInt
			}
			return cbr, nil, valToRet + int(int(cBuff[0])<<cBitSize)
		}
		valToRet += int(cBuff[0]&127) << cBitSize
		cBitSize += 7
	}
	return cbr, overflowError, valToRet
}
