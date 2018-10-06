package crawl

import (
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/integrii/flaggy"
	"golang.org/x/net/html"
)

var cfgFn string

func init() {
	flaggy.String(&cfgFn, "c", "config", "config file of crawler.")
}

type crawl struct {
	url      *url.URL
	doc      *goquery.Document
	config   *Config
	title    string
	location string
	content  []*position
}

func New(url *url.URL, doc *goquery.Document) (*crawl, error) {
	// TODO: select and init config
	config, err := readConfigFile(cfgFn)
	if err != nil {
		return nil, err
	}
	return &crawl{url: url, doc: doc, config: config}, nil
}

func (c *crawl) Title() string {
	var title string
	config := c.config
	c.doc.Find(config.Title.S.Path).Each(func(i int, center *goquery.Selection) {
		if i == config.Title.Index {
			title = strings.TrimSpace(center.Text())
		}
	})
	return title
}

func (c *crawl) contentProcess(processor func(int, *goquery.Selection)) {
	config := c.config
	c.doc.Find(config.Content.S.Path).Contents().Each(
		func(i int, content *goquery.Selection) {
			processor(i, content)
			content.Children().Each(func(i int, content *goquery.Selection) {
				processor(i, content)
			})
		})
}

func (c *crawl) Location() string {
	if c.location != "" {
		return c.location
	}
	config := c.config
	locationRegexp := regexp.MustCompile(config.Content.Location)
	c.contentProcess(func(i int, content *goquery.Selection) {
		if c.location != "" {
			return
		}
		match := locationRegexp.FindStringSubmatch(content.Text())
		if len(match) > 1 {
			c.location = match[1]
		}
	})
	return c.location
}

func (c *crawl) Content() []*position {
	if c.content != nil {
		return c.content
	}
	c.content = make([]*position, 0)
	config := c.config
	startRegexp := regexp.MustCompile("(?im)" + config.Content.Start)
	startFlag := false
	endFlag := false
	filters := make([]*regexp.Regexp, len(config.Content.Filter), len(config.Content.Filter))
	for idx, reg := range config.Content.Filter {
		filters[idx] = regexp.MustCompile("(?im)" + reg)
	}
	c.contentProcess(func(i int, content *goquery.Selection) {
		if endFlag {
			return
		}
		end := config.End.S
		if end.Path != "" {
			if content.Is(end.Path) {
				endFlag = true
				attr_pair := strings.Split(end.Attr, "==")
				if len(attr_pair) > 0 {
					attr_key := strings.TrimSpace(attr_pair[0])
					if attr_key != "" {
						attr_value := content.AttrOr(attr_key, "")
						if len(attr_pair) == 2 {
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
		if !startFlag {
			indexes := startRegexp.FindStringSubmatchIndex(content.Text())
			if len(indexes) > 0 {
				startFlag = true
			} else {
				return
			}
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
						imageUrl = c.url.ResolveReference(imageUrl)
						pos := position{Image: imageUrl.String()}
						if config.Content.Name < 0 {
							namePos := len(c.content) + config.Content.Name
							lastPos := len(c.content) - 1
							if namePos >= 0 && namePos < len(c.content) {
								if c.content[namePos].Image == "" && c.content[namePos].Name != "" {
									pos.Name = c.content[namePos].Name
								} else if c.content[lastPos].Image != "" && c.content[lastPos].Name != "" {
									pos.Name = c.content[lastPos].Name
								}
							}
						}
						c.content = append(c.content, &pos)
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
						if config.Content.Name > 0 {
							for _, pos := range c.content {
								if pos.Name == "" && pos.Image != "" {
									pos.Name = text
								}
							}
						}
						c.content = append(c.content, &position{Name: text})
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
					if config.Content.Name > 0 {
						for _, pos := range c.content {
							if pos.Name == "" && pos.Image != "" {
								pos.Name = text
							}
						}
					}
					c.content = append(c.content, &position{Name: text})
				}
			default:
				// TODO raise exception
				log.Fatalf("Got unknown node type %d", node.Type)
				return
			}
		}
	})
	return c.content
}
