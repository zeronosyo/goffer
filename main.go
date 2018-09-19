package main

import (
  "fmt"
  "log"
  "sync"
  "time"
  "math/rand"
  "strings"
  "net/http"

  "github.com/PuerkitoBio/goquery"
)

func filter_whitespace(orig string) string {
  cleared_space := strings.Replace(orig, " ", "", -1)
  cleared_space_and_newline := strings.Replace(cleared_space, "\n", "", -1)
  return cleared_space_and_newline
}

func randint(min int, max int) int {
  return rand.Intn(max - min) + min
}

func crawl(url string) string {
  if url == "" {
    return ""
  }
  if !strings.HasPrefix(url, "http:") {
    if strings.HasPrefix(url, "..") {
      url = "http://dotproducer.kan-be.com" + url[2:len(url) - 1]
    } else {
      url = "http://dotproducer.kan-be.com/seiti/" + url
    }
  }
  // FIXME
  time.Sleep(time.Duration(randint(10, 5000)) * time.Millisecond)
  res, err := http.Get(url)
  if err != nil {
    log.Fatal(err)
  }
  defer res.Body.Close()
  if res.StatusCode != 200 {
    if res.StatusCode != 404 {
      log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
    }
    return "Page not found: " + url
  }

  doc, err := goquery.NewDocumentFromReader(res.Body)
  if err != nil {
    log.Fatal(err)
  }

  title := ""
  doc.Find("body center").Each(func(i int, center *goquery.Selection) {
    if i == 0 {
      title = filter_whitespace(center.Text())
    }
  })
  return title
}

func main() {
  res, err := http.Get("http://dotproducer.kan-be.com/seiti/seiti.html")
  if err != nil {
    log.Fatal(err)
  }
  defer res.Body.Close()
  if res.StatusCode != 200 {
    log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
  }

  doc, err := goquery.NewDocumentFromReader(res.Body)
  if err != nil {
    log.Fatal(err)
  }

  var wait_group sync.WaitGroup

  doc.Find("body > div.rightrepeat > div.leftrepeat > div > center").Each(func(i int, center *goquery.Selection) {
    center_text := filter_whitespace(center.Text())
    center.Find("a").Each(func(j int, a *goquery.Selection) {
      content := a.AttrOr("href", "")
      wait_group.Add(1)
      go func(i, j int, center_text, content string) {
        defer wait_group.Done()
        fmt.Printf("Review (%d,%d): %s - %s(%s)\n", i, j, center_text, content, crawl(content))
      }(i, j, center_text, content)
    })
    //
  })
  wait_group.Wait()
}
