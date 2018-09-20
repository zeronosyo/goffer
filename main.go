package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// CrawlContent represent content crawl from network.
type CrawlContent struct {
	title string
	url   string
}

func filterWhitespace(orig string) string {
	clearedSpace := strings.Replace(orig, " ", "", -1)
	clearedSpaceNewline := strings.Replace(clearedSpace, "\n", "", -1)
	return clearedSpaceNewline
}

// RandInt generate random int in range [min, max].
func RandInt(min int, max int) int {
	return rand.Intn(max-min) + min
}

func crawl(url string) (*CrawlContent, error) {
	content := &CrawlContent{}
	if url == "" {
		return nil, errors.New("No url")
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
	time.Sleep(time.Duration(RandInt(10, 5000)) * time.Millisecond)
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		if res.StatusCode != 404 {
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}
		return nil, errors.New("Page not found: " + content.url)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("body center").Each(func(i int, center *goquery.Selection) {
		if i == 0 {
			content.title = filterWhitespace(center.Text())
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

	doc.Find("body > div.rightrepeat > div.leftrepeat > div > center").Each(
		func(i int, center *goquery.Selection) {
			centerText := filterWhitespace(center.Text())
			for len(centerText) < 22 {
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
						fmt.Printf(
							"Review %s(%02d,%02d): Got Error - %s\n",
							centerText, i, j, err,
						)
						return
					}
					fmt.Printf(
						"Review %s(%02d,%02d): %s(%s)\n",
						centerText, i, j, crawlContent.title, crawlContent.url,
					)
				}(i, j, centerText, content)
			})
			//
		})
	waitGroup.Wait()
}
