package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"path"
)

func (app *App) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
	}

	vars := mux.Vars(r)

	// Get this meme
	meme, err := app.getMemeById(vars["id"])
	if err != nil {
		panic(err)
	}

	// Remove this dank meme
	filepath := path.Join(app.cfg.UploadsDir, meme.Filename)
	err = os.Remove(filepath)
	if err != nil {
		panic(err)
	}

	_, err = app.db.Exec("DELETE FROM memes WHERE id = ?", meme.Id)
	if err != nil {
		panic(err)
	}

	err = app.idx.Delete(fmt.Sprintf("%v", meme.Id))
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
