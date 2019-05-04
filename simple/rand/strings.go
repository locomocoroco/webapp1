package rand

import (
	"crypto/rand"
	"encoding/base64"
)

const RemToBytes = 32

func Bytes(n int) ([]byte, error) {
	out := make([]byte, n)
	_, err := rand.Read(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func NBytes(base64string string) (int, error) {
	r, err := base64.URLEncoding.DecodeString(base64string)
	if err != nil {
		return 0, err
	}
	return len(r), nil
}

func String(nBytes int) (string, error) {
	bytes, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
func RememberToken() (string, error) {
	return String(RemToBytes)
}
