package crawl

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/htmlindex"
	"gopkg.in/yaml.v2"

	"github.com/zeronosyo/goffer/exc"
	"github.com/zeronosyo/goffer/redis"
	"github.com/zeronosyo/goffer/utils"
)

type Selector struct {
	Path string `yaml:"path"`
	Attr string `yaml:"attr"`
	Tag  string `yaml:"tag"`
}

type element struct {
	S Selector `yaml:"selector"`
}

type Title struct {
	element `yaml:",inline"`
	Index   int `yaml:"index"`
}

type Content struct {
	element   `yaml:",inline"`
	Image     Selector `yaml:"image"`
	Start     string   `yaml:"start"`
	Locations []string `yaml:"locations"`
	Name      int      `yaml:"name"`
	Filter    []string `yaml:"filter"`
}

type End struct {
	element `yaml:",inline"`
}

type Config struct {
	Url     string  `yaml:"url"`
	Title   Title   `yaml:"title"`
	Content Content `yaml:"content"`
	End     End     `yaml:"end"`
}

type position struct {
	Name  string
	Image string
}

func decodeResponse(body io.Reader, header http.Header) (*html.Node, error) {
	e, err := htmlindex.Get(utils.DetermineCharset(header.Get("content-type")))
	if err != nil {
		return nil, err
	}
	if name, _ := htmlindex.Name(e); name != "utf-8" {
		body = e.NewDecoder().Reader(body)
	}
	node, err := html.Parse(body)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func nodeToString(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}

func QueryUrl(url string) (*goquery.Document, error) {
	cacheBody, err := redis.GetCache(url)
	if err == nil && cacheBody != "" {
		body, err := html.Parse(strings.NewReader(cacheBody))
		if err == nil {
			return goquery.NewDocumentFromNode(body), nil
		}
	}
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		if res.StatusCode != 404 {
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}
		return nil, &exc.ErrPageNotFound
	}
	body, err := decodeResponse(res.Body, res.Header)
	if err != nil {
		return nil, err
	}
	redis.SetCache(url, nodeToString(body))
	return goquery.NewDocumentFromNode(body), nil
}

func readConfigFile(filename string) (*Config, error) {
	c := &Config{
		Title: Title{Index: 0},
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
