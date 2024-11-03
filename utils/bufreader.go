package utils

import "io"

type BufReader interface {
	io.Reader
	io.ByteReader
}
