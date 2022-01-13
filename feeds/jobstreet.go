package feeds

import (
	"fmt"

	"go-crawl/helper"
	"go-crawl/models"
	"go-crawl/repository"

	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

const (
	webPage = "https://www.jobstreet.vn/t%C3%ACmvi%E1%BB%87c"
)

func JobStreet(repo repository.Repository) {
	var wg sync.WaitGroup

	pipe := make(chan string)
	done := make(chan bool)
	go func() {
		for {
			url, more := <-pipe
			if more {
				count, err := repo.FindByUrl(url, "recruitment_jobstreet")
				if err != nil {
					fmt.Println(err)
				}
				if count == 0 {
					if errExtract := extractInfoJob(url, repo); errExtract != nil {
						fmt.Println(errExtract)
					}
				} else {
					fmt.Printf("Exists %s", url)
				}
			} else {
				fmt.Println("Extract all url")
				done <- true
				return
			}
		}
	}()

	wg.Add(2)

	go getUrlByProvince(pipe, &wg)
	go getUrlByCategory(pipe, &wg)

	go func() {
		wg.Wait()
		close(pipe)
	}()
	<-done
}

func extractInfoJob(url string, repo repository.Repository) error {
	var job models.Job

	c := colly.NewCollector()
	c.SetRequestTimeout(120 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println(err)
	})

	c.OnHTML(".jobresults .job-card", func(e *colly.HTMLElement) {
		job.Url = "https://www.jobstreet.vn" + e.ChildAttr("h3.job-title > a", "href")
		job.Title = e.ChildText("h3.job-title > a")
		job.Company = e.ChildText("span.job-company")
		job.Location = e.ChildText("span.job-location")

		c.Visit(e.Request.AbsoluteURL(job.Url))
		c.OnHTML("div[class=heading-xsmall]", func(e *colly.HTMLElement) {
			job.Site = e.ChildText("span.site")
			job.CreatedAt = e.ChildText("span.listed-date")
		})

		if job.Site == "TopCV" {
			job.Descript = ""
		} else {
			c.OnHTML("div[class=-desktop-no-padding-top]", func(e *colly.HTMLElement) {
				job.Descript = e.Text
			})
		}
	})

	// Save in to mongodb
	errSave := repo.Save(job, "recruitment_jobstreet")
	if errSave != nil {
		fmt.Println(errSave)
	}

	c.Visit(url)

	return nil
}

// Get all search url by province
func getUrlByProvince(pipe chan<- string, wg *sync.WaitGroup) error {
	defer wg.Done()

	doc, err := helper.GetNewDocument(webPage)
	if err != nil {
		return err
	}

	// Get all search urls by province
	doc.Find("div[id=browse-locations] a[href]").Each(func(index int, province *goquery.Selection) {
		href, _ := province.Attr("href")
		urlProvince := fmt.Sprintf("https://www.jobstreet.vn%s", href)

		// Get total page count of each url by province
		totalPage, err := getTotalPage(urlProvince)
		if err != nil {
			fmt.Println(err)
		}

		// Merge all url pages by province
		for page := 1; page <= totalPage; page++ {
			urlProvinceByPage := fmt.Sprintf("%s?p=%d", urlProvince, page)
			pipe <- urlProvinceByPage
		}
	})

	return nil
}

// Get all search url by category
func getUrlByCategory(pipe chan<- string, wg *sync.WaitGroup) error {
	defer wg.Done()
	doc, err := helper.GetNewDocument(webPage)
	if err != nil {
		return err
	}

	// Get all search urls by category
	doc.Find("div[id=browse-categories] a[href]").Each(func(index int, category *goquery.Selection) {
		href, _ := category.Attr("href")
		urlCategory := fmt.Sprintf("https://www.jobstreet.vn%s", href)

		docChild, err := helper.GetNewDocument(urlCategory)
		if err != nil {
			fmt.Println(err)
		}

		// Get all search urls by category child
		docChild.Find("div[id=browse-keywords] a[href]").Each(func(index int, key *goquery.Selection) {
			href, _ := key.Attr("href")
			urlCategoryChild := fmt.Sprintf("https://www.jobstreet.vn%s", href)

			// Get total page count of each url by category child
			totalPage, err := getTotalPage(urlCategoryChild)
			if err != nil {
				fmt.Println(err)
			}

			// Merge all url pages by category child
			for page := 1; page <= totalPage; page++ {
				urlCategoryChildByPage := fmt.Sprintf("%s?p=%d", urlCategoryChild, page)
				pipe <- urlCategoryChildByPage
			}
		})
	})

	return nil
}

// get total page count of each url
func getTotalPage(url string) (int, error) {
	var totalPage int
	doc, err := helper.GetNewDocument(url)
	if err != nil {
		return 0, err
	}

	pageStr := doc.Find("div.search-results-count strong:last-child").Text()
	if pageStr != "" {
		totalPage, err = strconv.Atoi(pageStr)
		if err != nil {
			return 0, err
		}
	}

	return totalPage, nil
}
