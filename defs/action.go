package defs

import "time"

type ActionInfo struct {
	ID         string `json:"id"`
	AgentID    string `json:"coreClientID"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Timestamp  int64  `json:"timestamp"`
	ExpireTime int64  `json:"expireTime"`
}

func (a *ActionInfo) IsExpired() bool {
	return a.ExpireTime <= time.Now().Unix()
}
