package hashhelper

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/zlojkota/YL-1/internal/collector"
	"sync"
)

type Hasher struct {
	Key    string
	keyMux sync.Mutex
}

func (hsh *Hasher) SetKey(key string) {
	hsh.keyMux.Lock()
	hsh.Key = key
	hsh.keyMux.Unlock()
}

func (hsh *Hasher) hash(src string) string {
	if hsh.Key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(hsh.Key))
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

func (hsh *Hasher) TestBatchHash(src []*collector.Metrics) bool {
	ret := true
	for _, val := range src {
		if hsh.Hash(val) != val.Hash {
			ret = false
		}
	}
	return ret
}
