package main

import (
	"math/rand"
	"time"

	"github.com/northbright/random"
)

func genSleepTime(ID string, beginTime, endTime time.Time) (time.Duration, error) {
	wakeupTime, err := random.Time(beginTime, endTime)
	if err != nil {
		return 0, err
	}

	return wakeupTime.Sub(time.Now()), nil
}

func generatePhaseTwoPrice(startPrice int64) int64 {
	var price int64

	t := time.Now()

	rand.Seed(t.UnixNano())
	x := rand.Intn(100)

	switch {
	case x < 10:
		price = startPrice
	case x >= 10 && x < 30:
		price = startPrice + 100
	case x >= 30 && x < 65:
		price = startPrice + 200
	default:
		price = startPrice + 300
	}
	return price
}
