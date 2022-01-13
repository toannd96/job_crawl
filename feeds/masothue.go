package feeds

import (
	"fmt"
	"sync"

	"go-crawl/helper"
	"go-crawl/models"
	"go-crawl/repository"

	"github.com/PuerkitoBio/goquery"
)

const (
	basePath     = "https://www.masothue.com"
	companyPath  = "/tra-cuu-ma-so-thue-theo-loai-hinh-doanh-nghiep"
	businessPath = "/tra-cuu-ma-so-thue-theo-nganh-nghe"
)

func NewCompany() *models.Company {
	return &models.Company{
		TaxInfo: make(map[string]string),
	}
}

func Masothue(repo repository.Repository) {
	var wg sync.WaitGroup

	pipe := make(chan string)
	done := make(chan bool)

	go func() {
		for {
			url, more := <-pipe
			if more {
				fmt.Println("Extract", url)

				if errExtract := extractCompanyInfo(url, repo); errExtract != nil {
					fmt.Println(errExtract)
				}
			} else {
				fmt.Println("Extract all url")
				done <- true
				return
			}
		}
	}()

	wg.Add(1)

	go getUrl(pipe, &wg)

	go func() {
		wg.Wait()
		close(pipe)
	}()
	<-done
}

func getUrl(pipe chan<- string, wg *sync.WaitGroup) error {
	defer wg.Done()

	for page := 1; page <= 20; page++ {
		url := fmt.Sprintf("%s%s?page=%d", basePath, businessPath, page)
		doc, err := helper.GetNewDocument(url)
		if err != nil {
			return err
		}

		doc.Find("table tbody").Each(func(index int, tableHtml *goquery.Selection) {
			tableHtml.Find("tr").Each(func(indexTr int, rowHtml *goquery.Selection) {
				rowHtml.Find("td:last-child a[href]").Each(func(ndexTd int, tableCell *goquery.Selection) {
					href, _ := tableCell.Attr("href")
					for page := 1; page <= 10; page++ {
						urlTypeCompany := fmt.Sprintf("%s%s?page=%d", basePath, href, page)

						docChild, _ := helper.GetNewDocument(urlTypeCompany)
						docChild.Find("div.tax-listing h3 a[href]").Each(func(index int, info *goquery.Selection) {
							href, _ := info.Attr("href")
							urlInfoCompany := fmt.Sprintf("%s%s", basePath, href)
							pipe <- urlInfoCompany
						})
					}
				})
			})
		})
	}

	return nil
}

func extractCompanyInfo(url string, repo repository.Repository) error {
	company := NewCompany()

	doc, err := helper.GetNewDocument(url)
	if err != nil {
		return err
	}

	// extract tax info
	doc.Find("table.table-taxinfo").Each(func(index int, tableTaxHtml *goquery.Selection) {
		tableTaxHtml.Find("th span.copy").Each(func(indexTr int, rowTaxHtml *goquery.Selection) {
			company.Name = rowTaxHtml.Text()
		})

		tableTaxHtml.Find("tbody tr").Each(func(indexTr int, rowTaxHtml *goquery.Selection) {
			row := make([]string, 0)
			rowTaxHtml.Find("td").Each(func(ndexTd int, tableCell *goquery.Selection) {
				row = append(row, tableCell.Text())
			})

			if len(row) != 1 {
				company.TaxInfo[row[0]] = row[1]
			}
		})

	})

	// extract type business
	doc.Find("table.table").Each(func(index int, tableBusinessHtml *goquery.Selection) {
		tableBusinessHtml.Find("tbody tr").Each(func(indexTr int, rowBusinessHtml *goquery.Selection) {
			row := make([]string, 0)
			rowBusinessHtml.Find("td").Each(func(ndexTd int, tableCell *goquery.Selection) {
				row = append(row, tableCell.Text())
			})
			business := models.Business{
				ID:     row[0],
				Carees: row[1],
			}
			company.Business = append(company.Business, business)

			err := repo.Save(company, "company_masothue")
			if err != nil {
				fmt.Println(err)
			}
		})
	})

	return nil
}
