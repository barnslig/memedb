package main

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"net/http"
	"path"
	"strings"
)

type SearchResult struct {
	Meme
	Filepath string
	Tags     []Tag
	Hit      *search.DocumentMatch
}

type SearchResults []SearchResult

func (app *App) SearchHandler(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	queryString := strings.Join(values["q"], " ")

	// Just show everything if the query string is empty
	if queryString == "" {
		queryString = "*"
	}

	query := bleve.NewQueryStringQuery(queryString)
	search := bleve.NewSearchRequest(query)

	results, err := app.idx.Search(search)
	if err != nil {
		panic(err)
	}

	searchResults := SearchResults{}
	for _, hit := range results.Hits {
		meme, tags, err := app.getMemeByIdWithTags(hit.ID)
		if err != nil {
			panic(err)
		}

		searchResults = append(searchResults, SearchResult{
			Meme:     meme,
			Filepath: path.Join(app.cfg.UploadsPath, meme.Filename),
			Tags:     tags,
			Hit:      hit,
		})
	}

	app.renderTemplate(w, r, "search.html", map[string]interface{}{
		"Query":        queryString,
		"BleveResults": results,
		"Results":      searchResults,
	})
	return

	app.renderTemplate(w, r, "search.html", nil)
}
