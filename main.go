package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/integrii/flaggy"

	"github.com/zeronosyo/goffer/crawl"
	"github.com/zeronosyo/goffer/exc"
)

var cfgDir string

func init() {
	flaggy.String(&cfgDir, "d", "config-dir", "config files directory of crawler.")
	flaggy.Parse()
	crawl.Init(cfgDir)
}

func crawlQuery(docUrl *url.URL) {
	doc, err := crawl.QueryUrl(docUrl.String())
	if err != nil {
		// log.Fatal(err)
		fmt.Printf("Query %s got error: %s\n", docUrl, err)
		return
	}
	c, err := crawl.New(docUrl, doc)
	if err != nil {
		fmt.Printf("Review Got Error - %s\n", err)
		return
	}
	if len(c.Page()) > 0 {
		fmt.Printf("RPage: %v, %d\n", docUrl, len(c.Page()))
		for _, p := range c.Page() {
			crawlQuery(p)
		}
		return
	}
	if c.Title() == "" || len(c.Locations()) == 0 || len(c.Content()) == 0 {
		return
	}
	var content string
	for idx, pos := range c.Content() {
		if pos.Image == "" {
			continue
		}
		content += fmt.Sprintf("%d: %s(%s),", idx, pos.Name, pos.Image)
	}
	fmt.Printf(
		"Review %s(%s) - (%s) - %s - (%s)\n",
		c.Title(), docUrl, c.Locations(), content, c.Page(),
	)
}

func main() {
	doc, err := crawl.QueryUrl("http://dotproducer.kan-be.com/seiti/seiti.html")
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
					crawlQuery(docUrl)
				}(href)
			})
		})
	waitGroup.Wait()
}
