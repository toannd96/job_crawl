package helper

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

// get html document from url
func GetNewDocument(url string) (*goquery.Document, error) {
	resp, err := Get(url)
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
