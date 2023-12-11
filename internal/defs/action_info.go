package defs

import (
	"time"
)

type ActionInfo struct {
	ID         string   `json:"id"`
	Status     int      `json:"status"`
	Messages   [][]byte `json:"messages"`
	ExpireTime int64    `json:"expireTime"`
}

func (a ActionInfo) IsExpired() bool {
	return a.ExpireTime <= time.Now().Unix()
}
