package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gocolly/colly"
)

type childData struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Link   string `json:"link"`
}

type data struct {
	Title   string      `json:"title"`
	Author  string      `json:"author"`
	Link    string      `json:"link"`
	Related []childData `json:"related"`
}

type dataJSON struct {
	// stores the json of the data struct in a list
	DataArr []data `json:"data"`
}

// main() contains code adapted from example found in Colly's docs:
// http://go-colly.org/docs/examples/basic/
func main() {
	// Instantiate default collector
	c := colly.NewCollector()

	webLink := "https://news.google.com/topics/CAAqJggKIiBDQkFTRWdvSUwyMHZNRFp0Y1RjU0FtVnVHZ0pWVXlnQVAB?hl=en-US&gl=US&ceid=US%3Aen"
	cardSelector := "div > div > main > c-wiz > div > div > main > div.lBwEZb.BL5WZb.GndZbb > div:nth-child(1)"
	// relatedSelector := "#tabCAQqKggAKiYICiIgQ0JBU0Vnb0lMMjB2TURadGNUY1NBbVZ1R2dKVlV5Z0FQAQ > div > div > main > c-wiz > div > div > main > div.lBwEZb.BL5WZb.GndZbb > div:nth-child(1) > div > div.SbNwzf"
	// selectorTest := "#tabCAQqKggAKiYICiIgQ0JBU0Vnb0lMMjB2TURadGNUY1NBbVZ1R2dKVlV5Z0FQAQ > div > div > main > c-wiz > div > div > main > div.lBwEZb.BL5WZb.GndZbb > div:nth-child(1) > div > div.SbNwzf > article:nth-child(1) > h4"
	// #tabCAQqKggAKiYICiIgQ0JBU0Vnb0lMMjB2TURadGNUY1NBbVZ1R2dKVlV5Z0FQAQ > div > div > main > c-wiz > div > div > main > div.lBwEZb.BL5WZb.GndZbb > div:nth-child(1) > div > div.SbNwzf > article:nth-child(1) > a
	var datalist []data
	var d data

	c.OnHTML(cardSelector, func(e *colly.HTMLElement) {
		var link string
		var title string
		var author string
		var relatedList []childData

		mainTitle := e.ChildText("div > article > h3 > a")
		mainLink := "https://news.google.com" + e.ChildAttr("div > a", "href")[1:]
		mainAuthor := e.ChildText("a + article > div.QmrVtf.RD0gLb > div > a")

		e.ForEach("div > div > article + div > article", func(_ int, h *colly.HTMLElement) {

			link = "https://news.google.com" + h.ChildAttr("a", "href")[1:]
			title = h.ChildText("h4 > a")
			author = h.ChildText("div > div > a")
			p := childData{Link: link, Title: title, Author: author}
			relatedList = append(relatedList, p)
		})
		// relatedDataStruct = append(relatedDataStruct, relatedList...)
		// jsonStruct, err := json.MarshalIndent(relatedDataStruct, "", "	")
		// if err != nil {
		// 	panic(err)
		// }
		d = data{Title: mainTitle, Link: mainLink, Author: mainAuthor, Related: relatedList}
		datalist = append(datalist, d)
	})

	// // Before making a request print "Visiting ..."
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

	// fmt.Println(len(bodyList))

	ls := dataJSON{DataArr: datalist}

	DataJSONarr, err := json.MarshalIndent(ls, "", "	")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("output.json", DataJSONarr, 0644)
	if err != nil {
		panic(err)
	}

}
