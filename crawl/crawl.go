package crawl

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var (
	configMap map[*regexp.Regexp]*Config
)

func init() {
	configMap = make(map[*regexp.Regexp]*Config)
}

func Init(cfgDir string) {
	filepath.Walk(cfgDir, loadConfig)
}

func loadConfig(path string, f os.FileInfo, err error) error {
	if err != nil {
		log.Fatalf("Scan directory %s got error: %v", path, err)
	}
	if !f.IsDir() {
		config, err := readConfigFile(path)
		if err != nil {
			return err
		}
		if config.Url != "" {
			configMap[regexp.MustCompile("(?im)"+config.Url)] = config
			log.Printf("loaded config: %v\n", path)
		}
	}
	return nil
}

type crawl struct {
	url         *url.URL
	doc         *goquery.Document
	config      *Config
	title       string
	locations   []string
	content     []*position
	startRegexp *regexp.Regexp
}

func New(docUrl *url.URL, doc *goquery.Document) (*crawl, error) {
	var config *Config
	for urlRegexp := range configMap {
		if len(urlRegexp.FindStringSubmatchIndex(docUrl.String())) > 0 {
			config = configMap[urlRegexp]
			break
		}
	}
	if config == nil {
		log.Fatalf("No such config file for url: %v", docUrl)
	}
	c := crawl{url: docUrl, doc: doc, config: config}
	if config.Content.Start != "" {
		c.startRegexp = regexp.MustCompile("(?im)" + config.Content.Start)
	}
	return &c, nil
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

func (c *crawl) contentProcess(processor func(int, *goquery.Selection) bool) {
	config := c.config
	continuous := true
	c.doc.Find(config.Content.S.Path).Contents().Each(
		func(i int, content *goquery.Selection) {
			if !continuous {
				return
			}
			continuous = processor(i, content)
			if !continuous {
				return
			}
			content.Children().Each(func(i int, content *goquery.Selection) {
				if !continuous {
					return
				}
				continuous = processor(i, content)
				if !continuous {
					return
				}
			})
			if !continuous {
				return
			}
		})
}

func (c *crawl) Locations() []string {
	if c.locations != nil {
		return c.locations
	}
	c.locations = make([]string, 0)
	locationRegexps := make([]*regexp.Regexp, 0)
	config := c.config
	for _, loc := range config.Content.Locations {
		locationRegexps = append(locationRegexps, regexp.MustCompile(loc))
	}
	c.contentProcess(func(i int, content *goquery.Selection) bool {
		for _, regexp := range locationRegexps {
			match := regexp.FindStringSubmatch(content.Text())
			if len(match) > 1 {
				c.locations = append(c.locations, match[1])
				break
			}
		}
		return true
	})
	return c.locations
}

// Ensure end elementif Return true otherwise return false
func (c *crawl) ensureEnd(content *goquery.Selection) bool {
	end := c.config.End.S
	path := end.Path
	attr_pair := strings.Split(end.Attr, "==")
	tag := end.Tag
	if path == "" && len(attr_pair) > 0 && tag == "" {
		// without end symbol
		return false
	}
	if path != "" && !content.Is(path) {
		return false
	}
	if len(attr_pair) > 0 {
		attr_key := strings.TrimSpace(attr_pair[0])
		if attr_key != "" {
			attr_value := content.AttrOr(attr_key, "")
			if len(attr_pair) == 2 {
				if attr_value != strings.TrimSpace(attr_pair[1]) {
					return false
				}
			} else if attr_value == "" {
				return false
			}
		}
	}
	if tag != "" {
		if content.Nodes[0].Type != html.ElementNode || content.Nodes[0].Data != tag {
			return false
		}
	}
	return true
}

// Ensure start element if Return true otherwise return false
func (c *crawl) ensureStart(content *goquery.Selection) bool {
	if c.startRegexp == nil {
		return true
	}
	indexes := c.startRegexp.FindStringSubmatchIndex(content.Text())
	return len(indexes) > 0
}

func (c *crawl) Content() []*position {
	if c.content != nil {
		return c.content
	}
	c.content = make([]*position, 0)
	config := c.config
	started := false
	filters := make([]*regexp.Regexp, len(config.Content.Filter), len(config.Content.Filter))
	for idx, reg := range config.Content.Filter {
		filters[idx] = regexp.MustCompile("(?im)" + reg)
	}
	c.contentProcess(func(i int, content *goquery.Selection) bool {
		if c.ensureEnd(content) {
			// finish traverse content
			return false
		}
		if !started && !c.ensureStart(content) {
			// skip to next iterate
			return true
		}
		// started
		started = true
		for _, node := range content.Nodes {
			switch {
			case node.Type == html.ElementNode && node.Data == config.Content.Image.Tag:
				// NOTE process image node
				imageSrc := content.AttrOr("src", "")
				if imageSrc != "" {
					imageUrl, err := url.Parse(imageSrc)
					if err != nil {
						return true
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
			case (node.Type == html.ElementNode && node.Data != config.Content.Image.Tag) || node.Type == html.TextNode:
				// NOTE process text node
				text := strings.TrimSpace(content.Text())
				for _, matcher := range filters {
					if len(matcher.FindStringSubmatchIndex(text)) > 0 {
						text = ""
						break
					}
				}
				if text != "" {
					if len(c.content) >= i {
						for i := config.Content.Name; i > 0; i-- {
							if c.content[len(c.content)-i].Image != "" {
								// 有图片数据, skip
								break
							}
							if i > 1 {
								// 还有未判断元素, 继续迭代
								continue
							}
							// 从第(len(c.content)-config.Content.Name)个元素到最后一个元素都是文字
							// 则给前面的图片数据添加上当前元素的文字信息
							for _, pos := range c.content {
								if pos.Name == "" && pos.Image != "" {
									pos.Name = text
								}
							}
						}
					}
					c.content = append(c.content, &position{Name: text})
				}
			default:
				// TODO raise exception
				log.Fatalf("Got unknown node type %d", node.Type)
				return false
			}
		}
		return true
	})
	return c.content
}
