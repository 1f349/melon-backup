package comm

import (
	int_byte_utils "github.com/1f349/int-byte-utils"
	"io"
)

type ServiceList struct {
	List []string
}

func (s *ServiceList) WriteTo(w io.Writer) (n int64, err error) {
	bw, err := int_byte_utils.WriteIntAsBytes(len(s.List), w)
	if err != nil {
		return int64(bw), err
	}
	for _, v := range s.List {
		cbw, err := int_byte_utils.WriteIntAsBytes(len(v), w)
		bw += cbw
		if err != nil {
			return int64(bw), err
		}
		cbw, err = w.Write([]byte(v))
		bw += cbw
		if err != nil {
			return int64(bw), err
		}
	}
	return int64(bw), nil
}

func (s *ServiceList) ReadFrom(r io.Reader) (n int64, err error) {
	br, err, sz := int_byte_utils.ReadIntFromBytes(r)
	if err != nil {
		return int64(br), err
	}
	s.List = make([]string, sz)
	for i := 0; i < sz; i++ {
		cbr, err, csz := int_byte_utils.ReadIntFromBytes(r)
		br += cbr
		if err != nil {
			return int64(br), err
		}
		cbuff := make([]byte, csz)
		cbr, err = io.ReadFull(r, cbuff)
		br += cbr
		if err != nil {
			return int64(br), err
		}
		s.List[i] = string(cbuff)
	}
	return int64(br), err
}
