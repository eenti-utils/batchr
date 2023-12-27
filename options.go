package batchr

import "time"

type Opts struct {
	PollingInterval    time.Duration
	NumChecksAfterStop int
}
