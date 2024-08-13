package utils

import (
	"errors"
	"time"
)

type ThroughputLimiter struct {
	throughput uint
	locks      chan *LimiterLock
}

type LimiterLock struct {
	limiter *ThroughputLimiter
}

var (
	ErrTimeout     = errors.New("timeout")
	ErrInvalidLock = errors.New("invalid lock")
)

func NewThroughputLimiter(throughput uint) *ThroughputLimiter {
	limiter := &ThroughputLimiter{
		throughput: throughput,
	}

	limiter.locks = make(chan *LimiterLock, throughput)
	for range throughput {
		limiter.locks <- &LimiterLock{limiter: limiter}
	}

	return limiter
}

func (l *ThroughputLimiter) Acquire(timeout time.Duration) (*LimiterLock, error) {
	expireTimer := time.After(timeout)
	select {
	case lock := <-l.locks:
		return lock, nil
	case <-expireTimer:
		return nil, ErrTimeout
	}
}

func (l *ThroughputLimiter) Release(lock *LimiterLock) error {
	if lock.limiter != l {
		return ErrInvalidLock
	}

	l.locks <- lock
	return nil
}
