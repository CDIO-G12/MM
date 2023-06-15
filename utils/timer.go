package utils

import (
	"time"

	log "github.com/s00500/env_logger"
)

type TimerType struct {
	startTime time.Time
}

func NewTimer() TimerType {
	return TimerType{}
}

func (t *TimerType) Start() {
	log.Info("Timer started")
	t.startTime = time.Now()
}

func (t *TimerType) Now() time.Duration {
	return time.Since(t.startTime)
}

func (t *TimerType) Left() time.Duration {
	return time.Until(t.startTime.Add(time.Minute * 8))
}

func (t *TimerType) LeftPercent() int {
	return int(100 * (t.Left().Seconds() / (8 * time.Minute.Seconds())))
}
