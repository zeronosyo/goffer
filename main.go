package main

import (
	"fmt"
	"log"

	"github.com/integrii/flaggy"

	"github.com/zeronosyo/goffer/crawl"
)

var cfgDir string

func init() {
	flaggy.String(&cfgDir, "d", "config-dir", "config files directory of crawler.")
	flaggy.Parse()
	crawl.Init(cfgDir)
}

func main() {
	crawlContent, err := kanbe()
	if err != nil {
		log.Fatal(err)
	}
	for _, cc := range crawlContent {
		var content string
		for idx, pos := range cc.Content {
			if pos.Image == "" {
				continue
			}
			content += fmt.Sprintf("%d: %s(%s),", idx, pos.Name, pos.Image)
		}
		log.Printf("Review %s - %s(%s): %v", cc.Url, cc.Title, cc.Locations, content)
	}
}
