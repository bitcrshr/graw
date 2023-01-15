package grawcore

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/markphelps/optional"
)

type RateLimiter struct {
	Remaining            optional.Float64
	NextRequestTimestamp time.Time
	ResetTimestamp       time.Time
	Used                 optional.Int
}

func NewRateLimiter(
	remaining optional.Float64,
	nextRequestTimestamp time.Time,
	resetTimestamp time.Time,
	used optional.Int,
) *RateLimiter {
	return &RateLimiter{
		Remaining:            remaining,
		NextRequestTimestamp: nextRequestTimestamp,
		ResetTimestamp:       resetTimestamp,
		Used:                 used,
	}
}

func (r RateLimiter) Delay() {
	if r.NextRequestTimestamp.IsZero() {
		return
	}

	sleepSeconds := r.NextRequestTimestamp.UTC().Sub(time.Now().UTC())
	if sleepSeconds.Seconds() <= 0.0 {
		return
	}

	message := fmt.Sprintf("Sleeping %0.2f seconds prior to call", sleepSeconds.Seconds())
	log.Println(message)
	time.Sleep(sleepSeconds)
}

func (r RateLimiter) Update(responseHeaders *map[string]string) {
	headers := *responseHeaders

	if headers["x-ratelimit-remaining"] == "" {
		if r.Remaining.Present() {
			remaining := r.Remaining.MustGet()
			r.Remaining.Set(remaining - 1)

			used := r.Used.OrElse(0)
			used += 1
		}

		return
	}

	now := time.Now().UTC()

	secondsToReset := headers["x-ratelimit-reset"]
	remaining := headers["x-ratelimit-remaining"]
	used := headers["x-ratelimit-used"]

	if remainingFloat, err := strconv.ParseFloat(remaining, 64); err == nil {
		r.Remaining.Set(remainingFloat)
	}

	if usedInt, err := strconv.Atoi(used); err == nil {
		r.Used.Set(usedInt)
	}

	if secondsToResetInt, err := strconv.Atoi(secondsToReset); err == nil {
		resetTimestamp := now.Add(time.Duration(secondsToResetInt * 1e9))
		r.ResetTimestamp = resetTimestamp
	
	if r.Remaining.MustGet() <= 0 {
		r.NextRequestTimestamp = r.ResetTimestamp
		return
	}

	// TODO set next request timestamp
}
