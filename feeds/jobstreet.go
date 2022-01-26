package feeds

import (
	"fmt"

	"go-crawl/common"
	"go-crawl/models"
	"go-crawl/repository"

	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

const (
	jobStreetBasePath   = "https://www.jobstreet.vn"
	jobStreetCareerPath = "/t%C3%ACmvi%E1%BB%87c"
)

func JobStreet(repo repository.Repository) {
	var wg sync.WaitGroup

	pipe := make(chan string)
	done := make(chan bool)
	go func() {
		for {
			url, more := <-pipe
			if more {
				fmt.Println("Visit", url)
				if errExtract := extractRecruitmentJobStreet(url, repo); errExtract != nil {
					fmt.Println(errExtract)
				}

			} else {
				fmt.Println("Extract all url jobstreet")
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

func extractRecruitmentJobStreet(url string, repo repository.Repository) error {
	var recruitment models.Recruitment

	c := colly.NewCollector()
	c.SetRequestTimeout(120 * time.Second)

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println(err)
	})

	c.OnHTML(".jobresults .job-card", func(e *colly.HTMLElement) {
		urlChild := fmt.Sprintf("%s%s", jobStreetBasePath, e.ChildAttr("h3.job-title > a", "href"))
		urlResult := common.RemoveCharacterInString(urlChild, "?")

		count, err := repo.FindByUrl(urlResult, "recruitment_jobstreet")
		if err != nil {
			fmt.Println(err)
		}
		if count == 0 {
			fmt.Println("Extract", urlResult)
			recruitment.Url = urlResult
			recruitment.Title = e.ChildText("h3.job-title > a")
			recruitment.Company = e.ChildText("span.job-company")
			recruitment.Location = e.ChildText("span.job-location")

			c.Visit(e.Request.AbsoluteURL(recruitment.Url))
			c.OnHTML("div[class=heading-xsmall]", func(e *colly.HTMLElement) {
				recruitment.Site = e.ChildText("span.site")
				recruitment.CreatedAt = e.ChildText("span.listed-date")
			})

			if recruitment.Site == "TopCV" {
				recruitment.Descript = ""
			} else {
				c.OnHTML("div[class=-desktop-no-padding-top]", func(e *colly.HTMLElement) {
					recruitment.Descript = e.Text
				})
			}
		} else {
			fmt.Printf("Exists %s\n", urlResult)
		}

		// Save in to mongodb
		errSave := repo.Save(recruitment, "recruitment_jobstreet")
		if errSave != nil {
			fmt.Println(errSave)
		}
	})

	c.Visit(url)

	return nil
}

// Get all search url by province
func getUrlByProvince(pipe chan<- string, wg *sync.WaitGroup) error {
	defer wg.Done()

	url := fmt.Sprintf("%s%s", jobStreetBasePath, jobStreetCareerPath)
	doc, err := common.GetNewDocument(url)
	if err != nil {
		return err
	}

	// Get all search urls by province
	doc.Find("div[id=browse-locations] a[href]").Each(func(index int, province *goquery.Selection) {
		href, _ := province.Attr("href")
		urlProvince := fmt.Sprintf("%s%s", jobStreetBasePath, href)

		// Get total page count of each url by province
		totalPage, err := getTotalPageJobStreet(urlProvince)
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

	url := fmt.Sprintf("%s%s", jobStreetBasePath, jobStreetCareerPath)
	doc, err := common.GetNewDocument(url)
	if err != nil {
		return err
	}

	// Get all search urls by category
	doc.Find("div[id=browse-categories] a[href]").Each(func(index int, category *goquery.Selection) {
		href, _ := category.Attr("href")
		urlCategory := fmt.Sprintf("%s%s", jobStreetBasePath, href)

		docChild, err := common.GetNewDocument(urlCategory)
		if err != nil {
			fmt.Println(err)
		}

		// Get all search urls by category child
		docChild.Find("div[id=browse-keywords] a[href]").Each(func(index int, key *goquery.Selection) {
			href, _ := key.Attr("href")
			urlCategoryChild := fmt.Sprintf("%s%s", jobStreetBasePath, href)

			// Get total page count of each url by category child
			totalPage, err := getTotalPageJobStreet(urlCategoryChild)
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
func getTotalPageJobStreet(url string) (int, error) {
	var totalPage int
	doc, err := common.GetNewDocument(url)
	if err != nil {
		return 0, err
	}

	numberPageStr := doc.Find("div.search-results-count strong:last-child").Text()
	if numberPageStr != "" {
		totalPage, err = strconv.Atoi(numberPageStr)
		if err != nil {
			return 0, err
		}
	}

	return totalPage, nil
}
