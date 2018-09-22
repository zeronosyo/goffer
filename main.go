package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
  "net/url"

  "github.com/integrii/flaggy"
	"github.com/PuerkitoBio/goquery"

  "github.com/zeronosyo/goffer/exc"
  "github.com/zeronosyo/goffer/crawl"
)

func main() {
  flaggy.Parse()
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
				go func(i, j int, center_text, href string) {
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
          doc, err := crawl.QueryUrl(docUrl.String())
          if err != nil {
            // log.Fatal(err)
            fmt.Printf("Query %s got error: %s", docUrl, err)
            return
          }
          crawlContent, err := crawl.Crawl(docUrl, doc)
					if err != nil {
						// fmt.Printf("Review %s(%02d,%02d): Got Error - %s\n", centerText, i, j, err)
						return
					}
          var content string
          for idx, pos := range crawlContent.Content {
            content += fmt.Sprintf("%d: %s(%s),", idx, pos.Name, pos.Image)
          }
					fmt.Printf(
						"Review %9s(%02d,%02d): %s(%s) - (%s) - %s\n",
						centerText, i, j, crawlContent.Title,
            docUrl, crawlContent.Location, content,
					)
				}(i, j, centerText, href)
			})
		})
	waitGroup.Wait()
}
