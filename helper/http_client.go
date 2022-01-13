package helper

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
)

const maxRetry = 3 * time.Minute

// Get http request basic
func get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Get http request with backoff retry
func Get(url string) (*http.Response, error) {
	var err error
	var resp *http.Response
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = maxRetry
	bo.MaxElapsedTime = maxRetry
	for {
		resp, err = get(url)
		if err == nil {
			break
		}
		fmt.Println("BackOff retry")
		d := bo.NextBackOff()
		if d == backoff.Stop {
			fmt.Println("Retry time out")
			break
		}
		fmt.Println("Retry in ", d)
		time.Sleep(d)
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}
