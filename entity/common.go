package entity

import "time"

var DefaultLocation *time.Location

func init() {
	wib, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}

	DefaultLocation = wib
}
