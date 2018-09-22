package crawl

import (
	"log"
  "io"
  "io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
  "golang.org/x/net/html"
  "golang.org/x/text/encoding/htmlindex"
  "gopkg.in/yaml.v2"

	"github.com/zeronosyo/goffer/utils"
  "github.com/zeronosyo/goffer/exc"
)

type position struct {
  Name string
  Image string
}

// CrawlContent represent content crawl from network.
type CrawlContent struct {
	Title string
	Url   string
  Content []*position
  Location string
}

type selector struct {
  Path string `yaml:"path"`
  Attr string `yaml:"attr"`
  Tag string `yaml:"tag"`
}

type title struct {
  Selector selector `yaml:"selector"`
  Index int `yaml:"index"`
}

type content struct {
  Selector selector `yaml:"selector"`
  Start string `yaml:"start"`
  Location string `yaml:"location"`
  Name int `yaml:"name"`
  Image selector `yaml:"image"`
  End selector `yaml:"end"`
  Filter []string `yaml:"filter"`
}

type conf struct {
  Title title `yaml:"title"`
  Content content `yaml:"content"`
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

func QueryUrl(url string) (*goquery.Document, error) {
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
  return goquery.NewDocumentFromNode(body), nil
}

func readConfigFile(filename string) (*conf, error) {
  c := &conf{
    Title: title{Index: 0},
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
