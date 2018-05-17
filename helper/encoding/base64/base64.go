package base64

import (
	"bytes"
	"encoding/base64"
	"io"
)

func URLEncoding(src []byte) ([]byte, error) {
	return encoding(base64.URLEncoding, src)
}

func URLDecoding(src []byte) ([]byte, error) {
	return decoding(base64.URLEncoding, src)
}

func StdEncoding(src []byte) ([]byte, error) {
	return encoding(base64.StdEncoding, src)
}

func StdDecoding(src []byte) ([]byte, error) {
	return decoding(base64.StdEncoding, src)
}

func RawStdEncoding(src []byte) ([]byte, error) {
	return encoding(base64.RawStdEncoding, src)
}

func RawStdDecoding(src []byte) ([]byte, error) {
	return decoding(base64.RawStdEncoding, src)
}

func encoding(encoding *base64.Encoding, src []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	encoder := base64.NewEncoder(encoding, buf)
	encoder.Write(src)
	err := encoder.Close()
	return buf.Bytes(), err
}

func decoding(encoding *base64.Encoding, src []byte) ([]byte, error) {
	source := bytes.NewBuffer(src)
	decoder := base64.NewDecoder(encoding, source)
	raw := bytes.NewBuffer(nil)
	_, err := io.Copy(raw, decoder)
	return raw.Bytes(), err
}
