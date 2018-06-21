package main

import (
	"fmt"
	"github.com/opennota/phash"
	"github.com/speps/go-hashids"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func (app *App) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parse the form data
		r.ParseMultipartForm(100000)
		file, handler, err := r.FormFile("file")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// Make sure there are tags
		tagStrings := parseTags(strings.Join(r.Form["tags"], ","))
		if len(tagStrings) == 0 {
			panic("SPECIFY TAGS YOU *******!!!")
		}

		// Create a new database entry with a salt
		rand.Seed(time.Now().UnixNano())
		salt := int64(rand.Intn(1000))
		res, err := app.db.Exec("INSERT INTO memes (salt) VALUES (?)", salt)
		if err != nil {
			panic(err)
		}

		// Get the Id of the new database entry
		id, err := res.LastInsertId()
		if err != nil {
			panic(err)
		}
		memeId := fmt.Sprintf("%v", id)

		// Create the hash ID
		hd := hashids.NewData()
		hd.Salt = string(salt)
		h, err := hashids.NewWithData(hd)
		if err != nil {
			panic(err)
		}
		hash, err := h.EncodeInt64([]int64{id})
		if err != nil {
			panic(err)
		}

		// Save the file
		filepath := path.Join(app.cfg.UploadsDir, hash+path.Ext(handler.Filename))
		f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		io.Copy(f, file)

		// Calculate the phash
		phash, err := phash.ImageHashDCT(filepath)
		if err != nil {
			panic(err)
		}

		// Update the table to match the final file name
		_, err = app.db.Exec("UPDATE memes SET path = ?, phash = ? WHERE id = ?", path.Base(filepath), fmt.Sprintf("%d", phash), id)
		if err != nil {
			panic(err)
		}

		// Set the tags
		err = app.setMemeTags(memeId, tagStrings)
		if err != nil {
			panic(err)
		}

		// Update the search index
		app.reindexMeme(memeId)

		http.Redirect(w, r, fmt.Sprintf("/duplicates/%v", id), http.StatusFound)
	}

	app.renderTemplate(w, r, "upload.html", nil)
}
