package common

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
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
func getWithRetry(url string) (*http.Response, error) {
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

// get html document from url
func GetNewDocument(url string) (*goquery.Document, error) {
	resp, err := getWithRetry(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	return doc, nil
}
