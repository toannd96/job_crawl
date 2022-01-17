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
	"go-crawl/repository"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

const (
	itviecBasePath = "https://itviec.com"
	itviecJobsPath = "/it-jobs"
	itviecSignin   = "/sign_in"

	googleSignin = "https://accounts.google.com"
)

func loginTask() (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("remote-debugging-port", "9222"),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	// Login google
	googleTask(ctx)

	// Login itviec
	itviecTask(ctx)

	return ctx, cancel
}

func ItViec(repo repository.Repository) {
	// var urls []string

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
					// urls = append(urls, url)

					if errExtract := ExtractRecruitmentTask(ctx, url, repo); errExtract != nil {
						fmt.Println(errExtract)
					}
				} else {
					fmt.Printf("Exists %s\n", url)
				}
			} else {
				// fmt.Println("Extract all url itviec", len(urls))
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

	// totalPage, err := getTotalPageItViec()
	// if err != nil {
	// 	return err
	// }

	urls := []string{
		"https://itviec.com/it-jobs/backend-developer-java-spring-mysql-ftech-co-ltd-0345",
		"https://itviec.com/it-jobs/devops-engineer-ci-cd-engineer-forix-4048",
	}

	for _, url := range urls {
		pipe <- url

		// for page := 1; page <= 1; page++ {
		// 	url := fmt.Sprintf("%s%s?page=%d", itViecBasePath, itViecJobsPath, page)
		// 	doc, err := common.GetNewDocument(url)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	doc.Find("h3.title a[href]").Each(func(index int, content *goquery.Selection) {
		// 		href, _ := content.Attr("href")
		// 		urlRecruitment := common.RemoveCharacterInString(fmt.Sprintf("%s%s", itViecBasePath, href), "?")
		// 		pipe <- urlRecruitment
		// 	})
		// }
	}
	return nil
}

func ExtractRecruitmentTask(ctx context.Context, url string, repo repository.Repository) error {
	// var recruitment models.Recruitment

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

			doc.Find("div.job-details__overview div.svg-icon__text").Each(func(index int, info *goquery.Selection) {
				salary := info.Text()
				fmt.Println(salary)
			})

			return nil
		}),
	}

	if err := chromedp.Run(ctx, task); err != nil {
		log.Fatal(err)
	}
	return nil
}

func itviecTask(ctx context.Context) {
	url := fmt.Sprintf("%s%s", itviecBasePath, itviecSignin)
	task := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(3 * time.Second),
	}

	if err := chromedp.Run(ctx, task); err != nil {
		log.Fatal(err)
	}
}

// Login with email
// func itviecTask(ctx context.Context) {
// 	url := fmt.Sprintf("%s%s", itviecBasePath, itviecSignin)
// 	email := "//*[@id='user_email']"
// 	password := "//*[@id='user_password']"
// 	label := "//*[@id='container']/div[2]/div/div[2]/form/div[4]/div/div/div/iframe"
// 	button := "//*[@id='container']/div[2]/div/div[2]/form/div[5]/div/button"
// 	task := chromedp.Tasks{
// 		chromedp.Navigate(url),

// 		chromedp.SendKeys(email, ""),
// 		chromedp.Sleep(3 * time.Second),

// 		chromedp.SendKeys(password, ""),
// 		chromedp.Sleep(3 * time.Second),

// 		chromedp.Click(label),
// 		chromedp.Sleep(5 * time.Second),

// 		chromedp.Click(button),
// 	}

// 	if err := chromedp.Run(ctx, task); err != nil {
// 		log.Fatal(err)
// 	}
// }

func googleTask(ctx context.Context) {
	email := "//*[@id='identifierId']"
	password := "//*[@id='password']/div[1]/div/div[1]/input"
	buttonEmailNext := "//*[@id='identifierNext']/div/button"
	buttonPasswordNext := "//*[@id='passwordNext']/div/button/span"

	task := chromedp.Tasks{
		chromedp.Navigate(googleSignin),
		chromedp.SendKeys(email, ""),
		chromedp.Sleep(3 * time.Second),

		chromedp.Click(buttonEmailNext),
		chromedp.Sleep(3 * time.Second),

		chromedp.SendKeys(password, ""),
		chromedp.Sleep(3 * time.Second),

		chromedp.Click(buttonPasswordNext),
	}

	if err := chromedp.Run(ctx, task); err != nil {
		log.Fatal(err)
	}
}
