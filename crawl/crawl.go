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
  crawlContent.Content = make([]*position, 0)
  startRegexp := regexp.MustCompile("(?im)" + config.Content.Start)
  startFlag := false
  endFlag := false
  locationRegexp := regexp.MustCompile(config.Content.Location)
  filters := make([]*regexp.Regexp, len(config.Content.Filter), len(config.Content.Filter))
  for idx, reg := range config.Content.Filter {
    filters[idx] = regexp.MustCompile("(?im)" + reg)
  }
  var contentProcess func(int, *goquery.Selection)
  contentProcess = func(i int, content *goquery.Selection) {
    if endFlag {
      return
    }
    end := config.End.Selector
    if end.Path != "" {
      if content.Is(end.Path) {
        endFlag = true
        attr_pair := strings.Split(end.Attr, "==")
        if len(attr_pair) > 0 {
          attr_key := strings.TrimSpace(attr_pair[0])
          if attr_key != "" {
            attr_value := content.AttrOr(attr_key, "")
            if len(attr_pair) == 2  {
              if attr_value != strings.TrimSpace(attr_pair[1]) {
                endFlag = false
              }
            } else if attr_value == "" {
              endFlag = false
            }
          }
        }
        tag := end.Tag
        if tag != "" {
          if content.Nodes[0].Type != html.ElementNode || content.Nodes[0].Data != tag {
            endFlag = false
          }
        }
      }
    }
    if endFlag {
      return
    }
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
      } else {
        return
      }
    }
    posImage := 0
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
            if len(crawlContent.Content) > 0 {
              if config.Content.Name < 0 {
                if crawlContent.Content[len(crawlContent.Content)-1].Image == "" {
                  posImage = len(crawlContent.Content)
                }
              } else if config.Content.Name > 0 {
                if crawlContent.Content[len(crawlContent.Content)-1].Image != "" {
                  posImage = len(crawlContent.Content)
                }
              }
            }
            namePos := posImage + config.Content.Name
            var pos position
            if namePos >= 0 && namePos < len(crawlContent.Content) {
              pos = position{Name: crawlContent.Content[namePos].Name, Image: imageUrl.String()}
            } else {
              pos = position{Image: imageUrl.String()}
            }
            crawlContent.Content = append(crawlContent.Content, &pos)
          }
        default:
          text := strings.TrimSpace(content.Text())
          for _, matcher := range filters {
            if len(matcher.FindStringSubmatchIndex(text)) > 0 {
              text = ""
              break
            }
          }
          if text != "" {
            if config.Content.Name == len(crawlContent.Content) - posImage {
              for _, pos := range crawlContent.Content {
                if pos.Name == "" {
                  pos.Name = text
                }
              }
            }
            crawlContent.Content = append(crawlContent.Content, &position{Name: text})
          }
        }
      case html.TextNode:
        text := strings.TrimSpace(content.Text())
        for _, matcher := range filters {
          if len(matcher.FindStringSubmatchIndex(text)) > 0 {
            text = ""
            break
          }
        }
        if text != "" {
          if config.Content.Name == len(crawlContent.Content) - posImage {
            for _, pos := range crawlContent.Content {
              if pos.Name == "" {
                pos.Name = text
              }
            }
          }
          crawlContent.Content = append(crawlContent.Content, &position{Name: text})
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
