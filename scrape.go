package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gocolly/colly"
)

type relatedData struct {
	RelatedTitle  string `json:"title"`
	RelatedAuthor string `json:"author"`
	RelatedLink   string `json:"link"`
	RelatedDate   string `json:"date"`
}

type data struct {
	Title   string        `json:"title"`
	Author  string        `json:"author"`
	Link    string        `json:"link"`
	Date    string        `json:"date"`
	Related []relatedData `json:"related"`
}

var dataJSON map[string]interface{}

// main() contains code adapted from example found in Colly's docs:
// http://go-colly.org/docs/examples/basic/
func main() {
	// Instantiate default collector
	// e := echo.New()
	c := colly.NewCollector()
	webLink := "https://news.google.com/"

	var d data
	var topicLink string
	var topicName string
	linkSelector := "div.gb_zc > div.gb_Ec > div > c-wiz > div > div"
	dataJSON = make(map[string]interface{})

	c.OnHTML(linkSelector, func(k *colly.HTMLElement) {
		// fmt.Println(k.Text)
		y := colly.NewCollector()
		var datalist []data
		cardSelector := "div > div > main > c-wiz > div > div > main > div:first-child"

		k.ForEach("a[href*='./topics']", func(_ int, s *colly.HTMLElement) {
			topicName = k.ChildText("a[href*='./topics'] > div.e20EGc")
			topicLink = "https://news.google.com" + k.ChildAttr("a[href*='./topics']", "href")[1:]

			y.OnHTML(cardSelector, func(s *colly.HTMLElement) {
				s.ForEach("div.NiLAwe", func(_ int, e *colly.HTMLElement) {
					var relatedList []relatedData
					mainTitle := e.ChildText("div > article > h3 > a")
					mainLink := "https://news.google.com" + e.ChildAttr("div > a", "href")
					if e.ChildAttr("div > a", "href") == "" {
						mainLink = "https://news.google.com"
					} else {
						mainLink = "https://news.google.com" + e.ChildAttr("div > a", "href")[1:]
					}
					mainAuthor := e.ChildText("article:only-of-type > div.QmrVtf.RD0gLb > div > a")
					datePosted := e.ChildAttr("article > div.QmrVtf.RD0gLb > div > time", "datetime")[:10]

					e.ForEach("div > div > article + div > article", func(_ int, h *colly.HTMLElement) {
						relatedLink := "https://news.google.com" + h.ChildAttr("a", "href")[1:]
						relatedTitle := h.ChildText("h4 > a")
						relatedAuthor := h.ChildText("div > div > a")
						relatedDate := h.ChildAttr("div > div > time", "datetime")[:10]
						p := relatedData{RelatedLink: relatedLink, RelatedTitle: relatedTitle, RelatedAuthor: relatedAuthor, RelatedDate: relatedDate}
						relatedList = append(relatedList, p)
					})
					d = data{Title: mainTitle, Link: mainLink, Author: mainAuthor, Related: relatedList, Date: datePosted}
					relatedList = nil
					datalist = append(datalist, d)
					dataJSON[topicName] = datalist
				})

			})
			y.OnRequest(func(r *colly.Request) {
				fmt.Println("Visiting", r.URL.String())
			})
			y.OnScraped(func(r *colly.Response) {
				fmt.Println("Finished", r.Request.URL)
			})
			y.OnError(func(_ *colly.Response, err error) {
				log.Println("Something went wrong:", err)
			})
			y.Visit(topicLink)

		})
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {

		fmt.Println("Visiting", r.URL.String())
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
	})

	// Start scraping on https://hackerspaces.org
	c.Visit(webLink)
	json, err := json.MarshalIndent(dataJSON, "", "	")
	if err != nil {
		panic(err)
	}

	// ls := dataJSON{ScienceArr: datalist}

	// e.GET("/scrape", func(f echo.Context) error {
	// 	return f.JSON(http.StatusOK, ls)
	// })

	err = ioutil.WriteFile("output.json", json, 0644)
	if err != nil {
		panic(err)
	}

	// e.Logger.Fatal(e.Start(":1323"))
}
