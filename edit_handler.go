package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"path"
	"strings"
)

func (app *App) EditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	meme, err := app.getMemeById(vars["id"])
	if err != nil {
		panic(err)
	}

	tags, err := app.getTagsByMemeId(vars["id"])
	if err != nil {
		panic(err)
	}

	if r.Method == "POST" {
		r.ParseForm()

		memeId := fmt.Sprintf("%v", meme.Id)

		// Set new tags
		tagStrings := parseTags(strings.Join(r.Form["tags"], ","))
		app.setMemeTags(memeId, tagStrings)

		// Update the bleve search index
		app.reindexMeme(memeId)

		http.Redirect(w, r, fmt.Sprintf("/edit/%v", meme.Id), http.StatusFound)
		return
	}

	app.renderTemplate(w, r, "edit.html", map[string]interface{}{
		"Filepath": path.Join(app.cfg.UploadsPath, meme.Filename),
		"Filetype": path.Ext(meme.Filename),
		"Meme":     meme,
		"Tags":     tags,
	})
}
