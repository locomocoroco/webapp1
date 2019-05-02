package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

func NewHMAC(key string) HMAC {
	return HMAC{
		hmac: hmac.New(sha256.New, []byte(key)),
	}
}

type HMAC struct {
	hmac hash.Hash
}

func (h HMAC) Hash(in string) string {
	h.hmac.Reset()
	h.hmac.Write([]byte(in))
	return base64.URLEncoding.EncodeToString(h.hmac.Sum(nil))
}
