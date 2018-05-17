package gob

import (
	"bytes"
	"encoding/gob"
	. "goil/helper/encoding"
)

func init() {
	gob.Register(map[Void]Void{})
}

func EncodeMap(m map[Void]Void) ([]byte, error) {
	//register the type of Void
	for _, v := range m {
		gob.Register(v)
	}
	buf := bytes.NewBuffer(nil)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(m)
	return buf.Bytes(), err
}

func DecodeMap(buf []byte) (map[Void]Void, error) {
	reader := bytes.NewReader(buf)
	decoder := gob.NewDecoder(reader)
	m := make(map[Void]Void)
	err := decoder.Decode(&m)
	return m, err
}

type Wrapper struct {
	Val Void
}

func GobEncode(iface Void) ([]byte, error) {
	w := Wrapper{
		Val: iface,
	}

	gob.Register(w.Val)

	buf := bytes.NewBuffer(nil)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(w)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GobDecode(buf []byte) (Void, error) {
	w := Wrapper{}
	reader := bytes.NewReader(buf)
	decoder := gob.NewDecoder(reader)
	err := decoder.Decode(&w)
	if err != nil {
		return nil, err
	}
	return w.Val, nil
}
