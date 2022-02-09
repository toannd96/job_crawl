package feeds

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-crawl/common"
	"go-crawl/models"
	"go-crawl/repository"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

const (
	googleSignin = "https://accounts.google.com"

	itviecBasePath = "https://itviec.com"
	itviecJobsPath = "/it-jobs"
	itviecSignin   = "/sign_in"
)

func loginTask() (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		// chromedp.Flag("start-fullscreen", true),

		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("remote-debugging-port", "9222"),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	// login google
	googleTask(ctx)

	// login itviec with google
	itviecWithGoogleTask(ctx)

	return ctx, cancel
}

func ItViec(repo repository.Repository) {
	ctx, cancel := loginTask()

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

					if errExtract := ExtractItviecTask(ctx, url, repo); errExtract != nil {
						fmt.Println(errExtract)
					}
				} else {
					fmt.Printf("Exists %s\n", url)
				}
			} else {
				fmt.Println("Extract all url itviec")
				cancel()
				done <- true
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go getUrlItViec(pipe, &wg)

	go func() {
		wg.Wait()
		close(pipe)
	}()
	<-done
}

func getTotalPageItViec() (int, error) {
	url := fmt.Sprintf("%s%s", itviecBasePath, itviecJobsPath)
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
		url := fmt.Sprintf("%s%s?page=%d", itviecBasePath, itviecJobsPath, page)
		doc, err := common.GetNewDocument(url)
		if err != nil {
			return err
		}
		doc.Find("h3.title a[href]").Each(func(index int, content *goquery.Selection) {
			href, _ := content.Attr("href")
			urlRecruitment := common.RemoveCharacterInString(fmt.Sprintf("%s%s", itviecBasePath, href), "?")
			pipe <- urlRecruitment
		})
	}

	return nil
}

func ExtractItviecTask(ctx context.Context, url string, repo repository.Repository) error {
	var recruitment models.Recruitment
	recruitment.Url = url

	task := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			if err != nil {
				return err
			}
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(res))
			if err != nil {
				return err
			}

			doc.Find("div.job-details").Each(func(index int, body *goquery.Selection) {

				// title
				body.Find("h1.job-details__title").Each(func(index int, title *goquery.Selection) {
					recruitment.Title = title.Text()
				})

				// company
				body.Find("div.job-details__sub-title").Each(func(index int, company *goquery.Selection) {
					recruitment.Company = common.RemoveCharacterInString(company.Text(), "\n")
				})

				// skill
				body.Find("div.job-details__tag-list a.mkt-track span").Each(func(index int, skillKeyword *goquery.Selection) {
					recruitment.SkillKeyword = append(recruitment.SkillKeyword, skillKeyword.Text())
				})

				body.Find("div.job-details__overview div.svg-icon__text span").Each(func(index int, address *goquery.Selection) {
					if len(strings.Split(address.Text(), "\n")) == 2 {
						// address
						recruitment.Address = append(recruitment.Address, strings.Split(address.Text(), "\n")[0], strings.Split(address.Text(), "\n")[1])
					} else {
						// address
						recruitment.Address = append(recruitment.Address, strings.Split(address.Text(), "\n")[0])
					}

				})

				body.Find("div.job-details__overview div.svg-icon__text").Each(func(index int, info *goquery.Selection) {
					// salary
					recruitment.Salary = strings.Split(info.Text(), "\n")[0]
				})

				infoJobDescript := make([]string, 0)
				body.Find("div.job-details__paragraph ul li").Each(func(index int, details *goquery.Selection) {
					infoJobDescript = append(infoJobDescript, details.Text())
				})

				recruitment.Descript = strings.Join(infoJobDescript, "\n")
			})

			// save in to mongodb
			errSave := repo.Save(recruitment, "recruitment_itviec")
			if errSave != nil {
				fmt.Println(errSave)
			}

			return nil
		}),
	}

	if err := chromedp.Run(ctx, task); err != nil {
		fmt.Println(err)
	}
	return nil
}

func itviecWithGoogleTask(ctx context.Context) {
	url := fmt.Sprintf("%s%s", itviecBasePath, itviecSignin)

	task := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(2 * time.Second),
	}

	if err := chromedp.Run(ctx, task); err != nil {
		fmt.Println(err)
	}
}

func googleTask(ctx context.Context) {
	email := "//*[@id='identifierId']"
	password := "//*[@id='password']/div[1]/div/div[1]/input"
	buttonEmailNext := "//*[@id='identifierNext']/div/button"
	buttonPasswordNext := "//*[@id='passwordNext']/div/button/span"

	task := chromedp.Tasks{
		chromedp.Navigate(googleSignin),
		chromedp.SendKeys(email, ""),
		chromedp.Sleep(2 * time.Second),

		chromedp.Click(buttonEmailNext),
		chromedp.Sleep(2 * time.Second),

		chromedp.SendKeys(password, ""),
		chromedp.Sleep(2 * time.Second),

		chromedp.Click(buttonPasswordNext),
		chromedp.Sleep(3 * time.Second),
	}

	if err := chromedp.Run(ctx, task); err != nil {
		fmt.Println(err)
	}
}
