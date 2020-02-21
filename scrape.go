package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"

	"gopkg.in/gookit/color.v1"
)

// Stores the db value globally to be called
var dB *gorm.DB

// struct for the related google articles
type relatedArticle struct {
	gorm.Model
	RelatedTitle  string `json:"title"`
	RelatedTopic  string `json:"topic"`
	RelatedAuthor string `json:"author"`
	RelatedLink   string `json:"link"`
	RelatedDate   string `json:"date"`
}

// struct for the main articles
type article struct {
	gorm.Model
	Title   string           `json:"title"`
	Topic   string           `json:"topic"`
	Author  string           `json:"author"`
	Link    string           `json:"link"`
	Date    string           `json:"date"`
	Related []relatedArticle `json:"related"`
}

// struct for a pair of topics and a list of articles to be used with channels
type listPair struct {
	Topic       string
	ArticleList []article
}

// struct for a pair of links and topics to be used with channels
type linkPair struct {
	Topic string
	Link  string
}

// vars to store the number of links searched and the number of articles saved
var numOfLinks int
var numOfArticles int

// converts related articles to articles to be stored in the DB
func convertAndStore(related relatedArticle) {
	rArticle := article{Title: related.RelatedTitle, Topic: related.RelatedTopic, Author: related.RelatedAuthor, Link: related.RelatedLink, Date: related.RelatedDate}

	check := dB.NewRecord(&rArticle)
	if check == true {
		dB.Create(&rArticle)
	}

}

// scrapes all of the individual articles and their corresponding related articles
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
			if mainTitle != "" {
				numOfArticles++
			}
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
				if relatedTitle != "" {
					numOfArticles++
				}
				relatedAuthor := h.ChildText("div > div > a")
				relatedDate := h.ChildAttr("div > div > time", "datetime")[:10]
				p := relatedArticle{RelatedTopic: lpair.Topic, RelatedLink: relatedLink, RelatedTitle: relatedTitle, RelatedAuthor: relatedAuthor, RelatedDate: relatedDate}
				convertAndStore(p)
				relatedList = append(relatedList, p)
			})
			d = article{Topic: lpair.Topic, Title: mainTitle, Link: mainLink, Author: mainAuthor, Related: relatedList, Date: datePosted}

			check := dB.NewRecord(&d)
			if check == true {
				dB.Create(&d)
			}

			relatedList = nil
			articleList = append(articleList, d)
		})
	})

	y.Visit(lpair.Link)
	pair := listPair{Topic: lpair.Topic, ArticleList: articleList}
	return pair
}

// scrapes all of the links for the topics
func linkScrape() interface{} {
	defer timeSince(time.Now())

	c := colly.NewCollector()
	webLink := "https://news.google.com/"
	var topicLink string
	var topicName string
	// topicList = append(topicList, topicName)

	linkSelector := "div.gb_zc > div.gb_Ec > div > c-wiz > div > div"
	articleMap := make(map[string]interface{})

	links := make(chan linkPair, 100)
	results := make(chan listPair, 100)

	c.OnHTML(linkSelector, func(k *colly.HTMLElement) {
		k.ForEach("a[href*='./topics']", func(_ int, s *colly.HTMLElement) {
			topicName = k.ChildText("a[href*='./topics'] > div.e20EGc")
			topicLink = "https://news.google.com" + k.ChildAttr("a[href*='./topics']", "href")[1:]
			pair := linkPair{Topic: topicName, Link: topicLink}
			numOfLinks++
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

	close(links)
	for i := 1; i < 10; i++ {
		result := <-results
		articleMap[strings.ToLower(result.Topic)] = result.ArticleList
	}

	jsonarticle := jsonify(articleMap)

	writeJSONFile(jsonarticle)
	return articleMap
}

// converts the dictionary of topics and article lists to json to be served on the web
func jsonify(articleMap map[string]interface{}) []byte {
	json, err := json.MarshalIndent(articleMap, "", "	")
	if err != nil {
		panic(err)
	}
	return json
}

// writes a .json file that holds the json of the articles
func writeJSONFile(json []byte) {
	err := ioutil.WriteFile("output.json", json, 0644)
	if err != nil {
		panic(err)
	}
}

// worker gorountine function that scrapes each topic concurrently
func worker(links chan linkPair, results chan listPair) {
	for link := range links {
		results <- cardScrape(link)
	}
}

// prints out a line displaying the number of pages and articles scraped in the amount of time
func timeSince(start time.Time) {
	bold := color.Bold.Render
	success := color.Success.Render
	since := time.Since(start).Seconds()
	fmt.Printf("%s: Scraped %s pages and %d articles in %.2f seconds.\n", success("SUCCESS"), bold(numOfLinks), numOfArticles, since)
}

// starts the echo server
func startServer(jsonarticle interface{}) {
	e := echo.New()
	// e.Use(livereload.LiveReload())

	e.GET("/scrape", func(f echo.Context) error {
		return f.JSON(http.StatusOK, jsonarticle)
	})

	e.Logger.Fatal(e.Start(":5000"))

}

func main() {
	db, err := gorm.Open("sqlite3", "google-article.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	dB = db
	db.AutoMigrate(&article{})

	jsonData := linkScrape()

	startServer(jsonData)
}
