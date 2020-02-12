package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gocolly/colly"
)

type data struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Link   string `json:"link"`
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
	titleSelector := "#tabCAQqKggAKiYICiIgQ0JBU0Vnb0lMMjB2TURadGNUY1NBbVZ1R2dKVlV5Z0FQAQ > div > div > main > c-wiz > div > div > main > div.lBwEZb.BL5WZb.GndZbb > div > div > article > h3 > a"
	authorSelector := "#tabCAQqKggAKiYICiIgQ0JBU0Vnb0lMMjB2TURadGNUY1NBbVZ1R2dKVlV5Z0FQAQ > div > div > main > c-wiz > div > div > main > div.lBwEZb.BL5WZb.GndZbb > div > div > article > div.QmrVtf.RD0gLb > div.SVJrMe > a"

	var datalist []data
	var linkList []string
	var titleList []string
	var authorList []string
	var bodyList []string
	// On every a element which has href attribute call callback
	c.OnHTML(titleSelector, func(e *colly.HTMLElement) {
		link := "news.google.com" + e.Attr("href")
		linkList = append(linkList, link)
		titleList = append(titleList, e.Text)

		g := colly.NewCollector()
		g.OnHTML("p", func(e *colly.HTMLElement) {
			bodyList = append(bodyList, e.Text)
		})
		g.Visit(link)
		// fmt.Printf("Link found: %q -> %s\n", e.Text, link)

	})

	c.OnHTML(authorSelector, func(e *colly.HTMLElement) {
		authorList = append(authorList, e.Text)
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

	length := len(bodyList)

	for i := 0; i < length; i++ {
		d := data{Title: titleList[i], Author: authorList[i], Link: linkList[0][1:]}
		datalist = append(datalist, d)
	}

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
