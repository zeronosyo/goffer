package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/zeronosyo/goffer/utils"
  "github.com/zeronosyo/goffer/exc"
)

// CrawlContent represent content crawl from network.
type CrawlContent struct {
	title string
	url   string
}

func crawl(url string) (*CrawlContent, error) {
	content := &CrawlContent{}
  if url == "" {
    return nil, exc.RaiseHttpExc(404, "No url")
  }
	if !strings.HasPrefix(url, "http:") {
		if strings.HasPrefix(url, "..") {
			url = "http://dotproducer.kan-be.com" + url[2:len(url)-1]
		} else {
			url = "http://dotproducer.kan-be.com/seiti/" + url
		}
	}
	content.url = url
	// FIXME
	time.Sleep(time.Duration(utils.RandInt(10, 5000)) * time.Millisecond)
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		if res.StatusCode != 404 {
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}
		return nil, &exc.ErrPageNotFound
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("body center").Each(func(i int, center *goquery.Selection) {
		if i == 0 {
			content.title = strings.TrimSpace(center.Text())
		}
	})
	return content, nil
}

func main() {
	res, err := http.Get("http://dotproducer.kan-be.com/seiti/seiti.html")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var waitGroup sync.WaitGroup

	doc.Find("body center").Each(
		func(i int, center *goquery.Selection) {
			centerText := strings.TrimSpace(center.Text())
			if strings.HasSuffix(centerText, "New") {
				centerText = centerText[:len(centerText)-4]
			}
			centerText = strings.TrimSpace(centerText)
			for len(centerText) < 24 {
				centerText = "　" + centerText
			}
			if i == 0 {
				fmt.Printf("Title: %s\n", centerText)
				return
			}
			center.Find("a").Each(func(j int, a *goquery.Selection) {
				content := a.AttrOr("href", "")
				waitGroup.Add(1)
				go func(i, j int, center_text, content string) {
					defer waitGroup.Done()
					crawlContent, err := crawl(content)
					if err != nil {
						// fmt.Printf("Review %s(%02d,%02d): Got Error - %s\n", centerText, i, j, err)
						return
					}
					fmt.Printf(
						"Review %9s(%02d,%02d): %s(%s)\n",
						centerText, i, j, crawlContent.title, crawlContent.url,
					)
				}(i, j, centerText, content)
			})
		})
	waitGroup.Wait()
}
