package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"

	"github.com/zeronosyo/goffer/crawl"
	"github.com/zeronosyo/goffer/exc"
)

func kanbe() ([]*crawl.CrawlContent, error) {
	doc, err := crawl.QueryUrl("http://dotproducer.kan-be.com/seiti/seiti.html")
	if err != nil {
		return nil, err
	}

	var waitGroup sync.WaitGroup
	crawlContent := make([]*crawl.CrawlContent, 0)

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
				href := a.AttrOr("href", "")
				waitGroup.Add(1)
				go func(href string) {
					defer waitGroup.Done()
					if href == "" {
						log.Fatal(exc.RaiseHttpExc(404, "No url"))
					}
					docUrl, err := url.Parse(href)
					if err != nil {
						return
					}
					base, err := url.Parse("http://dotproducer.kan-be.com/seiti/")
					if err != nil {
						return
					}
					docUrl = base.ResolveReference(docUrl)
					cc, err := crawl.Crawling(docUrl)
					if err != nil {
						return
					}
					crawlContent = append(crawlContent, cc...)
				}(href)
			})
		})
	waitGroup.Wait()
	return crawlContent, nil
}
