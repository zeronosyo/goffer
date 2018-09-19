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

type CrawlContent struct {
	title string
	url   string
}

func filter_whitespace(orig string) string {
	cleared_space := strings.Replace(orig, " ", "", -1)
	cleared_space_and_newline := strings.Replace(cleared_space, "\n", "", -1)
	return cleared_space_and_newline
}

func randint(min int, max int) int {
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
	time.Sleep(time.Duration(randint(10, 5000)) * time.Millisecond)
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
			content.title = filter_whitespace(center.Text())
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
			center_text := filter_whitespace(center.Text())
			for len(center_text) < 22 {
				center_text = "　" + center_text
			}
			if i == 0 {
				fmt.Printf("Title: %s\n", center_text)
				return
			}
			center.Find("a").Each(func(j int, a *goquery.Selection) {
				content := a.AttrOr("href", "")
				waitGroup.Add(1)
				go func(i, j int, center_text, content string) {
					defer waitGroup.Done()
					crawl_content, err := crawl(content)
					if err != nil {
						fmt.Printf(
							"Review %s(%02d,%02d): Got Error - %s\n",
							center_text, i, j, err,
						)
						return
					}
					fmt.Printf(
						"Review %s(%02d,%02d): %s(%s)\n",
						center_text, i, j, crawl_content.title, crawl_content.url,
					)
				}(i, j, center_text, content)
			})
			//
		})
	waitGroup.Wait()
}
