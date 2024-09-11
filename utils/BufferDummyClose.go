package utils

import "bytes"

type BufferDummyClose struct {
	bytes.Buffer
}

func (b *BufferDummyClose) Close() error {
	return nil
}
