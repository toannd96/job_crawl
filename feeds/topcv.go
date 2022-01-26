package feeds

import (
	"fmt"
	"go-crawl/common"
	"go-crawl/models"
	"go-crawl/repository"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

const (
	topcvBasePath = "https://www.topcv.vn"
	topcvJobsPath = "/tim-viec-lam-moi-nhat"
)

func TopCV(repo repository.Repository) {
	pipe := make(chan string)
	done := make(chan bool)

	go func() {
		for {
			url, more := <-pipe
			if more {
				count, err := repo.FindByUrl(url, "recruitment_topcv")
				if err != nil {
					fmt.Println(err)
				}
				if count == 0 {
					fmt.Println("Extract", url)

					if errExtract := extractRecruitmentTopCV(url, repo); errExtract != nil {
						fmt.Println(errExtract)
					}
				} else {
					fmt.Printf("Exists %s\n", url)
				}
			} else {
				fmt.Println("Extract all url topcv")
				done <- true
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go GetUrlTopCV(pipe, &wg)

	go func() {
		wg.Wait()
		close(pipe)
	}()
	<-done
}

func GetUrlTopCV(pipe chan<- string, wg *sync.WaitGroup) error {
	defer wg.Done()

	for page := 1; page <= 400; page++ {
		url := fmt.Sprintf("%s%s?page=%d", topcvBasePath, topcvJobsPath, page)
		fmt.Println(url)
		doc, err := common.GetNewDocument(url)
		if err != nil {
			return err
		}
		doc.Find("h3.title a[href]").Each(func(index int, content *goquery.Selection) {
			href, _ := content.Attr("href")
			if !strings.Contains(href, "brand") && !strings.Contains(href, "26-tuoi") {
				pipe <- common.RemoveCharacterInString(href, "?")
			}
		})
	}
	return nil
}

func extractRecruitmentTopCV(url string, repo repository.Repository) error {
	var recruitment models.Recruitment
	recruitment.Url = url

	doc, err := common.GetNewDocument(url)
	if err != nil {
		return err
	}

	// info title
	doc.Find("div.box-info-job").Each(func(index int, infoTitleHtml *goquery.Selection) {
		// title job
		infoTitleHtml.Find("h1.job-title a[href]").Each(func(indexTr int, titleJobHtml *goquery.Selection) {
			recruitment.Title = titleJobHtml.Text()
		})
		infoTitleHtml.Find("h1.job-title").Each(func(indexTr int, titleJobHtml *goquery.Selection) {
			recruitment.Title = titleJobHtml.Text()
		})

		// name company
		infoTitleHtml.Find("div.company-title a[href]").Each(func(indexTr int, nameCompanyHtml *goquery.Selection) {
			recruitment.Company = nameCompanyHtml.Text()
		})

		// job deadline
		infoTitleHtml.Find("div.job-deadline").Each(func(indexTr int, jobDeadlineHtml *goquery.Selection) {
			recruitment.JobDeadline = strings.ReplaceAll(strings.ReplaceAll(jobDeadlineHtml.Text(), "\n", ""), "Hạn nộp hồ sơ:", "")
		})
	})

	// info common
	infoCommon := make([]string, 0)
	doc.Find("div.box-main div.box-item").Each(func(index int, infoCommonHtml *goquery.Selection) {
		infoCommonHtml.Find("div span").Each(func(indexTr int, infoHtml *goquery.Selection) {
			infoCommon = append(infoCommon, infoHtml.Text())
		})
	})
	recruitment.Salary = strings.ReplaceAll(infoCommon[0], "\n", "")
	recruitment.NumberRecruits = infoCommon[1]
	recruitment.WorkForm = infoCommon[2]
	recruitment.Rank = infoCommon[3]
	recruitment.Sex = infoCommon[4]
	recruitment.Experience = infoCommon[5]

	// info location
	infoLocation := make([]string, 0)
	doc.Find("div.box-address").Each(func(index int, locationHtml *goquery.Selection) {
		locationHtml.Find("div div.text-dark-gray").Each(func(indexTr int, addressHtml *goquery.Selection) {
			infoLocation = append(infoLocation, addressHtml.Text())
		})
	})
	recruitment.Location = strings.ReplaceAll(infoLocation[0], "- Khu vực:", "")
	if len(infoLocation) == 2 {
		recruitment.Address = strings.ReplaceAll(strings.ReplaceAll(infoLocation[1], "\n", ""), "-", "")
	}

	infoJobKeyword := make([]string, 0)
	infoSkillKeyword := make([]string, 0)
	doc.Find("div.box-keyword-job").Each(func(index int, keywordHtml *goquery.Selection) {
		// info keyword job
		keywordHtml.Find("div.keyword span a[href]").Each(func(indexTr int, keywordJobHtml *goquery.Selection) {
			infoJobKeyword = append(infoJobKeyword, keywordJobHtml.Text())
		})

		// info keyword skill
		keywordHtml.Find("div.skill span a[href]").Each(func(indexTr int, keywordSkillHtml *goquery.Selection) {
			infoSkillKeyword = append(infoSkillKeyword, keywordSkillHtml.Text())
		})
	})
	recruitment.JobKeyword = infoJobKeyword
	recruitment.SkillKeyword = infoSkillKeyword

	// job descript
	infoJobDescript := make([]string, 0)
	doc.Find("div.job-data").Each(func(index int, jobDescriptHtml *goquery.Selection) {
		jobDescriptHtml.Find("div.content-tab p").Each(func(indexTr int, desciptHtml *goquery.Selection) {
			infoJobDescript = append(infoJobDescript, desciptHtml.Text())
		})
	})
	recruitment.Descript = strings.Join(infoJobDescript, "")

	// Save in to mongodb
	errSave := repo.Save(recruitment, "recruitment_topcv")
	if errSave != nil {
		fmt.Println(errSave)
	}

	return nil
}
