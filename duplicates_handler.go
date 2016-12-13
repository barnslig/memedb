package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/opennota/phash"
	"net/http"
	"path"
	"sort"
)

type MemeWithDistance struct {
	Meme
	Distance int
	Filepath string
}

type MemesWithDistance []MemeWithDistance

func (mwd MemesWithDistance) Len() int {
	return len(mwd)
}

func (mwd MemesWithDistance) Less(i, j int) bool {
	return mwd[i].Distance < mwd[j].Distance
}

func (mwd MemesWithDistance) Swap(i, j int) {
	mwd[i], mwd[j] = mwd[j], mwd[i]
}

func (app *App) DuplicatesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Get the current meme
	currentMeme, err := app.getMemeById(vars["id"])
	if err != nil {
		panic(err)
	}

	// Get all other memes
	memes := MemesWithDistance{}
	rows, err := app.db.Query("SELECT id, salt, phash, created_at, path FROM memes WHERE id != ?", vars["id"])
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		meme := Meme{}
		err := rows.Scan(&meme.Id, &meme.Salt, &meme.PHash, &meme.CreatedAt, &meme.Filename)
		if err != nil {
			panic(err)
		}

		distance := phash.HammingDistance(currentMeme.PHash, meme.PHash)

		// Ignore everything over a certain threshold
		if distance >= 50 {
			continue
		}

		mwd := MemeWithDistance{
			Meme:     meme,
			Distance: distance,
			Filepath: path.Join(app.cfg.UploadsPath, meme.Filename),
		}

		memes = append(memes, mwd)
	}

	// Sort this stuff by it's distance
	sort.Sort(memes)

	// Redirect to the edit screen if there are no duplicates
	if memes.Len() == 0 {
		http.Redirect(w, r, fmt.Sprintf("/edit/%v", currentMeme.Id), http.StatusFound)
		return
	}

	app.renderTemplate(w, r, "duplicates.html", map[string]interface{}{
		"CurrentMemeFilepath": path.Join(app.cfg.UploadsPath, currentMeme.Filename),
		"CurrentMeme":         currentMeme,
		"Memes":               memes,
	})
}
