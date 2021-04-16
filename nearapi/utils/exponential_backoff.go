package utils

import (
	"errors"
	"fmt"
	"time"
)

type getResult func() (map[string]interface{}, error)

func ExponentialBackoff(
	startWaitTime, retryNumber int,
	waitBackoff float64,
	fn getResult,
) (map[string]interface{}, error) {
	waitTime := startWaitTime
	for i := 0; i < retryNumber; i++ {
		res, err := fn()
		if err == nil {
			return res, nil
		}
		// print error and continue
		fmt.Println(err)

		time.Sleep(time.Duration(waitTime) * time.Millisecond)
		waitTime = int(float64(waitTime) * waitBackoff)
	}

	return nil, errors.New("utils: exponential backoff failed")
}