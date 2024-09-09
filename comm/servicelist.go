package comm

import (
	"github.com/1f349/melon-backup/utils"
	"io"
)

type ServiceList struct {
	List []string
}

func (s *ServiceList) WriteTo(w io.Writer) (n int64, err error) {
	bw, err := utils.WriteCompressedInt(len(s.List), w)
	if err != nil {
		return int64(bw), err
	}
	for _, v := range s.List {
		cbw, err := utils.WriteCompressedInt(len(v), w)
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
	br, err, sz := utils.ReadCompressedInt(r)
	if err != nil {
		return int64(br), err
	}
	s.List = make([]string, sz)
	for i := 0; i < sz; i++ {
		cbr, err, csz := utils.ReadCompressedInt(r)
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
