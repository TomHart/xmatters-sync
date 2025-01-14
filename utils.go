package main

import (
	"log"
	"time"
)

func ParseTime(timeStr string) time.Time {
	layout := "2006-01-02T15:04:05Z" // Adjust the layout to match your time string format
	parsedTime, err := time.Parse(layout, timeStr)
	if err != nil {
		log.Fatalf("Unable to parse time: %v", err)
	}
	return parsedTime
}
