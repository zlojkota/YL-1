package hashhelper

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/zlojkota/YL-1/internal/collector"
	"sync"
)

type Hasher struct {
	key    string
	keyMux sync.Mutex
}

func (hsh *Hasher) SetKey(key string) {
	hsh.keyMux.Lock()
	hsh.key = key
	hsh.keyMux.Unlock()
}

func (hsh *Hasher) hash(src string) string {
	if hsh.key == "" {
		return ""
	}
	h := hmac.New(md5.New, []byte(hsh.key))
	h.Write([]byte(src))
	sign := h.Sum(nil)
	dst := make([]byte, hex.EncodedLen(len(sign)))
	hex.Encode(dst, sign)
	return string(dst)
}

func (hsh *Hasher) Hash(src *collector.Metrics) string {
	switch src.MType {
	case "gauge":
		return hsh.hash(fmt.Sprintf("%s:gauge:%f", src.ID, *src.Value))
	case "counter":
		return hsh.hash(fmt.Sprintf("%s:counter:%d", src.ID, *src.Delta))
	default:
		return "invalidType"
	}
}

func (hsh *Hasher) HashG(id string, val float64) string {
	return hsh.hash(fmt.Sprintf("%s:gauge:%f", id, val))
}

func (hsh *Hasher) HashC(id string, val int64) string {
	return hsh.hash(fmt.Sprintf("%s:counter:%d", id, val))
}

func (hsh *Hasher) TestHash(src *collector.Metrics) bool {
	if hsh.Hash(src) == src.Hash {
		return true
	} else {
		return false
	}
}
