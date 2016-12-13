package main

import (
	"net/http"
)

func (app *App) HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/search", http.StatusFound)
}
