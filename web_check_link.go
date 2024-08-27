package main

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// LinkIsLive is returning true if performing a HEAD request against
// the given URL within 1000ms is successful.
//
// If the request takes longer than 1000ms, but establishing a TCP
// connection works, then it also returns true.
//
// In all other cases, it returns false.
func LinkIsLive(link string) bool {
	httpClient := &http.Client{
		Timeout: 1000 * time.Millisecond,
	}
	response, err := httpClient.Head(link)
	if response != nil && response.StatusCode < 400 {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}
