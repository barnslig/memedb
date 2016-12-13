package main

import (
	"net/http"
)

func (app *App) TagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := app.getTags()
	if err != nil {
		panic(err)
	}

	app.renderTemplate(w, r, "tags.html", map[string]interface{}{
		"Tags": tags,
	})
}
