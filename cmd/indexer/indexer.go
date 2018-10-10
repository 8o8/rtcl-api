package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/34South/envr"
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/mikedonnici/pubmed"
)

const batchSize = 500
const backDays = 10

// Index specifies the index being created
type Index struct {
	Category string `json:"category"`
	Term     string `json:"term"`
	RelDate  int    `json:"reldate"`
}

// Article describes the storage format for an indexed article
type article struct {
	category string
	pubmed.Article
}

var algoliaClient algoliasearch.Client

func init() {

	envr.New("algoliaEnv", []string{
		"ALGOLIA_APP_ID",
		"ALGOLIA_ADMIN_KEY",
	}).Auto()

	algoliaClient = algoliasearch.NewClient(os.Getenv("ALGOLIA_APP_ID"), os.Getenv("ALGOLIA_ADMIN_KEY"))
}

func main() {
	add()
}

func add() {

	for category, term := range pubmedQueries {

		// Set up the query for the category
		p := pubmed.NewQuery(url.PathEscape(term))
		p.BackDays = backDays
		err := p.Search()
		if err != nil {
			log.Fatalln(err)
		}

		for i := 0; i < p.ResultCount; i++ {
			xa, err := p.Articles(i, batchSize)
			if err != nil {
				fmt.Println(err)
			}

			var indexObjects []algoliasearch.Object
			for _, a := range xa.Articles {
				rtcl := article{category, a}
				indexObjects = append(indexObjects, searchObject(rtcl))
			}

			index := algoliaClient.InitIndex("articles")
			res, err := index.AddObjects(indexObjects)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(res)

			i += batchSize
		}
	}
}

// convert a pubmed.Article to map[string]interface{} which is same as algoliasearch.Object
func searchObject(a article) algoliasearch.Object {

	so := algoliasearch.Object{}
	so["category"] = a.category
	so["objectID"] = strconv.Itoa(a.ID)
	so["title"] = a.Title
	so["url"] = a.URL
	so["keywords"] = a.Keywords
	so["pubName"] = a.Journal
	so["pubNameAbbr"] = a.JournalAbbrev
	so["pubVolume"] = a.Volume
	so["pubIssue"] = a.Issue
	so["pubPageRef"] = a.Pages
	so["pubTime"] = a.PubDate.Unix()
	so["pubDate"] = a.PubDate.Format("2006-01-02")

	// Must have a summary field for snippets to work
	so["summary"] = ""
	if len(a.Abstract) > 0 {
		so["summary"] = a.Abstract[0].Value
	}

	return so
}
