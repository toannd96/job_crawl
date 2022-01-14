package feeds

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"go-crawl/common"
	"go-crawl/repository"

	"github.com/PuerkitoBio/goquery"
)

const (
	itViecBasePath = "https://itviec.com"
	itViecJobsPath = "/it-jobs"
)

func ItViec(repo repository.Repository) {
	var urls []string
	var wg sync.WaitGroup

	pipe := make(chan string)
	done := make(chan bool)

	go func() {
		for {
			url, more := <-pipe
			if more {
				count, err := repo.FindByUrl(url, "recruitment_itviec")
				if err != nil {
					fmt.Println(err)
				}
				if count == 0 {
					fmt.Println("Extract", url)
					urls = append(urls, url)

					// if errExtract := extractRecruitmentItViec(url, repo); errExtract != nil {
					// 	fmt.Println(errExtract)
					// }
				} else {
					fmt.Printf("Exists %s\n", url)
				}
			} else {
				fmt.Println("Extract all url itviec", len(urls))
				done <- true
				return
			}
		}
	}()

	wg.Add(1)

	go getUrlItViec(pipe, &wg)

	go func() {
		wg.Wait()
		close(pipe)
	}()
	<-done
}

func getTotalPageItViec() (int, error) {
	url := fmt.Sprintf("%s%s", itViecBasePath, itViecJobsPath)
	doc, err := common.GetNewDocument(url)
	if err != nil {
		return 0, err
	}

	numberPageStr := doc.Find("ul.pagination li a").Text()
	common.RemoveCharacterInString(numberPageStr, ">")
	totalPageStr := strings.Split(common.RemoveCharacterInString(numberPageStr, ">"), "…")[1]

	totalPage, err := strconv.Atoi(totalPageStr)
	if err != nil {
		return 0, err
	}

	return totalPage, nil
}

func getUrlItViec(pipe chan<- string, wg *sync.WaitGroup) error {
	defer wg.Done()

	totalPage, err := getTotalPageItViec()
	if err != nil {
		return err
	}

	for page := 1; page <= totalPage; page++ {
		url := fmt.Sprintf("%s%s?page=%d", itViecBasePath, itViecJobsPath, page)
		doc, err := common.GetNewDocument(url)
		if err != nil {
			return err
		}
		doc.Find("h3.title a[href]").Each(func(index int, content *goquery.Selection) {
			href, _ := content.Attr("href")
			urlRecruitment := fmt.Sprintf("%s%s", itViecBasePath, href)
			pipe <- urlRecruitment
		})
	}

	return nil
}

func extractRecruitmentItViec(url string, repo repository.Repository) error {

	return nil
}
