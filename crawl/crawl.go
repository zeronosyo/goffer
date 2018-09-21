package crawl

import (
	"log"
  "net/url"
  "regexp"
	"strings"

  "github.com/integrii/flaggy"
	"github.com/PuerkitoBio/goquery"
  "golang.org/x/net/html"
)

var cfgFn string

func init() {
  flaggy.String(&cfgFn, "c", "config", "config file of crawler.")
}

func Crawl(base *url.URL, doc *goquery.Document) (*CrawlContent, error) {
	crawlContent := &CrawlContent{}
  config, err := readConfigFile(cfgFn)
  if err != nil {
    log.Fatal(err)
  }

	doc.Find(config.Title.Selector.Path).Each(func(i int, center *goquery.Selection) {
		if i == config.Title.Index {
			crawlContent.Title = strings.TrimSpace(center.Text())
		}
	})
  crawlContent.Content = make([]string, 0)
  startRegexp := regexp.MustCompile(config.Content.Start)
  startFlag := false
  locationRegexp := regexp.MustCompile(config.Content.Location)
  var contentProcess func(int, *goquery.Selection)
  contentProcess = func(i int, content *goquery.Selection) {
    if crawlContent.Location == "" {
      match := locationRegexp.FindStringSubmatch(content.Text())
      if len(match) > 1 {
        crawlContent.Location = match[1]
      }
    }
    if !startFlag {
      indexes := startRegexp.FindStringSubmatchIndex(content.Text())
      if len(indexes) > 0 {
        startFlag = true
      }
      return
    }
    for _, node := range content.Nodes {
      switch node.Type {
      case html.ElementNode:
        switch node.Data {
        case config.Content.Image.Tag:
          imageSrc := content.AttrOr("src", "")
          if imageSrc != "" {
            imageUrl, err := url.Parse(imageSrc)
            if err != nil {
              return
            }
            imageUrl = base.ResolveReference(imageUrl)
            crawlContent.Content = append(crawlContent.Content, imageUrl.String())
          }
        default:
          text := strings.TrimSpace(content.Text())
          if text != "" {
            crawlContent.Content = append(crawlContent.Content, text)
          }
        }
      case html.TextNode:
        text := strings.TrimSpace(content.Text())
        if text != "" {
          crawlContent.Content = append(crawlContent.Content, text)
        }
      default:
        // TODO raise exception
        log.Fatalf("Got unknown node type %d", node.Type)
        return
      }
    }
    content.Children().Each(func(i int, content *goquery.Selection) {
      contentProcess(i, content)
    })
  }
	doc.Find(config.Content.Selector.Path).Contents().Each(contentProcess)
	return crawlContent, nil
}
