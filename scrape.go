package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/labstack/echo"

	"gopkg.in/gookit/color.v1"
)

type relatedArticle struct {
	RelatedTitle  string `json:"title"`
	RelatedAuthor string `json:"author"`
	RelatedLink   string `json:"link"`
	RelatedDate   string `json:"date"`
}

type article struct {
	Title   string           `json:"title"`
	Author  string           `json:"author"`
	Link    string           `json:"link"`
	Date    string           `json:"date"`
	Related []relatedArticle `json:"related"`
}

type listPair struct {
	Topic       string
	ArticleList []article
}
type linkPair struct {
	Topic string
	Link  string
}

var articleMap map[string]interface{}
var num int

func cardScrape(lpair linkPair) listPair {
	cardSelector := "div > div > main > c-wiz > div > div > main > div:first-child"
	var relatedList []relatedArticle
	var datePosted string
	var mainTitle string
	var mainLink string
	var mainAuthor string
	var d article
	var articleList []article
	y := colly.NewCollector()
	y.OnHTML(cardSelector, func(s *colly.HTMLElement) {
		s.ForEach("div.NiLAwe", func(_ int, e *colly.HTMLElement) {

			mainTitle = e.ChildText("div > article > h3 > a")
			mainLink = e.ChildAttr("div > a", "href")
			if mainLink != "" {
				mainLink = "https://news.google.com" + mainLink[1:]
			} else {
				mainLink = "https://news.google.com"
			}
			mainAuthor = e.ChildText("article:only-of-type > div.QmrVtf.RD0gLb > div > a")
			datePosted = e.ChildAttr("article > div.QmrVtf.RD0gLb > div > time", "datetime")
			if e.ChildAttr("article > div.QmrVtf.RD0gLb > div > time", "datetime") != "" {
				datePosted = datePosted[:10]
			}

			e.ForEach("div > div > article + div > article", func(_ int, h *colly.HTMLElement) {
				relatedLink := "https://news.google.com" + h.ChildAttr("a", "href")[1:]
				relatedTitle := h.ChildText("h4 > a")
				relatedAuthor := h.ChildText("div > div > a")
				relatedDate := h.ChildAttr("div > div > time", "datetime")[:10]
				p := relatedArticle{RelatedLink: relatedLink, RelatedTitle: relatedTitle, RelatedAuthor: relatedAuthor, RelatedDate: relatedDate}
				relatedList = append(relatedList, p)
			})
			d = article{Title: mainTitle, Link: mainLink, Author: mainAuthor, Related: relatedList, Date: datePosted}
			relatedList = nil
			articleList = append(articleList, d)
		})
	})

	y.Visit(lpair.Link)
	pair := listPair{Topic: lpair.Topic, ArticleList: articleList}
	return pair
}

func linkScrape() interface{} {
	defer timeSince(time.Now())

	c := colly.NewCollector()
	webLink := "https://news.google.com/"
	var topicLink string
	var topicName string

	linkSelector := "div.gb_zc > div.gb_Ec > div > c-wiz > div > div"
	articleMap = make(map[string]interface{})

	links := make(chan linkPair, 100)
	results := make(chan listPair, 100)

	c.OnHTML(linkSelector, func(k *colly.HTMLElement) {
		k.ForEach("a[href*='./topics']", func(_ int, s *colly.HTMLElement) {
			topicName = k.ChildText("a[href*='./topics'] > div.e20EGc")
			topicLink = "https://news.google.com" + k.ChildAttr("a[href*='./topics']", "href")[1:]
			pair := linkPair{Topic: topicName, Link: topicLink}
			num++
			links <- pair
		})
	})

	c.Visit(webLink)
	go worker(links, results)
	go worker(links, results)
	go worker(links, results)
	go worker(links, results)
	go worker(links, results)
	go worker(links, results)
	go worker(links, results)
	go worker(links, results)
	go worker(links, results)

	close(links)
	for i := 1; i < 10; i++ {
		result := <-results
		articleMap[strings.ToLower(result.Topic)] = result.ArticleList
	}

	jsonarticle := jsonify(articleMap)

	writeJSONFile(jsonarticle)
	return articleMap
}

func jsonify(map[string]interface{}) []byte {
	json, err := json.MarshalIndent(articleMap, "", "	")
	if err != nil {
		panic(err)
	}
	return json
}

func writeJSONFile(json []byte) {
	err := ioutil.WriteFile("output.json", json, 0644)
	if err != nil {
		panic(err)
	}
}

func worker(links chan linkPair, results chan listPair) {
	for link := range links {
		results <- cardScrape(link)
	}
}

func timeSince(start time.Time) {
	bold := color.Bold.Render
	success := color.Success.Render
	since := time.Since(start).Seconds()
	fmt.Printf("%s: Scraped %s pages in %.2f seconds.\n", success("SUCCESS"), bold(num), since)
}
func startServer(jsonarticle interface{}) {
	e := echo.New()
	// e.Use(livereload.LiveReload())

	e.GET("/scrape", func(f echo.Context) error {
		return f.JSON(http.StatusOK, jsonarticle)
	})

	e.Logger.Fatal(e.Start(":5000"))

}

func main() {
	json := linkScrape()
	// something := make(map[string]interface{})
	startServer(json)
}
