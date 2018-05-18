package wtcPayload

import (
	"time"
)

type WtcPayload struct {
	Id          int
	Name        string
	Ip          string
	Ts          time.Time
	Hashrate    int
	Peercount   int
	BlockNumber int
}
