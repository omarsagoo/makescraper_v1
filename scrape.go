package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gocolly/colly"
	"github.com/labstack/echo/v4"
)

type childData struct {
	Title  string `json:"relatedTitle"`
	Author string `json:"relatedAuthor"`
	Link   string `json:"relatedLink"`
	Date   string `json:"relatedDate"`
}

type data struct {
	Title   string      `json:"relatedTitle"`
	Author  string      `json:"relatedAuthor"`
	Link    string      `json:"relatedLink"`
	Date    string      `json:"Date"`
	Related []childData `json:"related"`
}

type dataJSON struct {
	// stores the json of the data struct in a list
	ScienceArr []data `json:"science"`
}

// main() contains code adapted from example found in Colly's docs:
// http://go-colly.org/docs/examples/basic/
func main() {
	// Instantiate default collector
	e := echo.New()
	c := colly.NewCollector()

	webLink := "https://news.google.com/topics/CAAqJggKIiBDQkFTRWdvSUwyMHZNRFp0Y1RjU0FtVnVHZ0pWVXlnQVAB?hl=en-US&gl=US&ceid=US%3Aen"
	cardSelector := "div > div > main > c-wiz > div > div > main > div:first-child"
	var datalist []data
	var d data
	c.OnHTML(cardSelector, func(b *colly.HTMLElement) {

		b.ForEach("div.NiLAwe", func(_ int, e *colly.HTMLElement) {
			var relatedList []childData

			mainTitle := e.ChildText("div > article > h3 > a")
			mainLink := "https://news.google.com" + e.ChildAttr("div > a", "href")[1:]
			mainAuthor := e.ChildText("article:only-of-type > div.QmrVtf.RD0gLb > div > a")
			datePosted := e.ChildAttr("article > div.QmrVtf.RD0gLb > div > time", "datetime")[:10]

			e.ForEach("div > div > article + div > article", func(_ int, h *colly.HTMLElement) {

				relatedLink := "https://news.google.com" + h.ChildAttr("a", "href")[1:]
				relatedTitle := h.ChildText("h4 > a")
				relatedAuthor := h.ChildText("div > div > a")
				relatedDate := h.ChildAttr("div > div > time", "datetime")[:10]
				p := childData{Link: relatedLink, Title: relatedTitle, Author: relatedAuthor, Date: relatedDate}
				relatedList = append(relatedList, p)
			})
			d = data{Title: mainTitle, Link: mainLink, Author: mainAuthor, Related: relatedList, Date: datePosted}
			relatedList = nil
			datalist = append(datalist, d)
		})
	})

	// div:empty ~ div:has(~ div:empty)

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

	// linkSelector := "header > div.gb_zc > div.gb_Ec > div > c-wiz"
	// c.OnHTML(linkSelector, func(k *colly.HTMLElement) {
	// 	k.ForEach("div:empty ~ div:has(~ div:empty)", func(_ int, s *colly.HTMLElement) {
	// 		fmt.Println(s.ChildText("a > div:last-child > span"))
	// 	})
	// })
	// Start scraping on https://hackerspaces.org
	c.Visit(webLink)

	ls := dataJSON{ScienceArr: datalist}

	// e.POST("/scrape", func(m echo.Context) error {
	// 	if err := m.Bind(ls); err != nil {
	// 		return err
	// 	}
	// 	return m.JSON(http.StatusCreated, ls)
	// })
	e.GET("/scrape", func(f echo.Context) error {
		return f.JSON(http.StatusOK, ls)
	})

	DataJSONarr, err := json.MarshalIndent(ls, "", "	")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("output.json", DataJSONarr, 0644)
	if err != nil {
		panic(err)
	}

	e.Logger.Fatal(e.Start(":1323"))
}
