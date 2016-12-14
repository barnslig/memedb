package main

import (
	"fmt"
	"path"
	"time"
)

func (app *App) Migrate() {
	_, err := app.db.Exec(`
CREATE TABLE IF NOT EXISTS memes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	salt INTEGER NOT NULL,
	phash TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	path TEXT
);
CREATE INDEX IF NOT EXISTS memes_index ON memes(id, salt);

CREATE TABLE IF NOT EXISTS memes_tags (
	meme_id INTEGER NOT NULL,
	tag_id INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS memes_tags_index ON memes_tags(meme_id, tag_id);
CREATE INDEX IF NOT EXISTS tags_memes_index ON memes_tags(tag_id, meme_id);
CREATE UNIQUE INDEX IF NOT EXISTS memes_tags_unique_index ON memes_tags(meme_id, tag_id);

CREATE TABLE IF NOT EXISTS tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	tag TEXT UNIQUE NOT NULL
);
CREATE INDEX IF NOT EXISTS tags_index ON tags(tag);
	`)

	if err != nil {
		panic(err)
	}
}

type Meme struct {
	Id        int64
	Salt      int64
	PHash     uint64
	CreatedAt time.Time
	Filename  string
}

func (app *App) getMemeById(id string) (meme Meme, err error) {
	err = app.db.
		QueryRow("SELECT id, salt, phash, created_at, path FROM memes WHERE id = ?", id).
		Scan(&meme.Id, &meme.Salt, &meme.PHash, &meme.CreatedAt, &meme.Filename)
	return
}

func (app *App) getMemeByIdWithTags(id string) (meme Meme, tags []Tag, err error) {
	meme, err = app.getMemeById(id)
	if err != nil {
		return
	}

	tags, err = app.getTagsByMemeId(id)

	return
}

func (app *App) setMemeTags(id string, tags []string) (err error) {
	// Remove all old tag relations
	_, err = app.db.Exec("DELETE FROM memes_tags WHERE meme_id = ?", id)
	if err != nil {
		return
	}

	for _, tag := range tags {
		// Make sure the tag exists within the database
		_, err = app.db.Exec("INSERT OR IGNORE INTO tags (tag) VALUES (?)", tag)
		if err != nil {
			return
		}

		// Make sure the relation exists
		_, err = app.db.Exec("INSERT OR IGNORE INTO memes_tags (meme_id, tag_id) VALUES (?, (SELECT id FROM tags WHERE tag = ?))", id, tag)
		if err != nil {
			return
		}
	}

	return
}

type Tag struct {
	Id        int64
	CreatedAt time.Time
	Tag       string
}

func (app *App) getTags() (tags []Tag, err error) {
	rows, err := app.db.Query("SELECT id, created_at, tag FROM tags")
	if err != nil {
		return
	}

	for rows.Next() {
		tag := Tag{}
		err = rows.Scan(&tag.Id, &tag.CreatedAt, &tag.Tag)
		if err != nil {
			return
		}

		tags = append(tags, tag)
	}

	return
}

func (app *App) getTagsByMemeId(id string) (tags []Tag, err error) {
	query := `
SELECT tags.id, tags.created_at, tags.tag
FROM tags
INNER JOIN memes_tags ON tags.id = memes_tags.tag_id
WHERE memes_tags.meme_id = ?`

	rows, err := app.db.Query(query, id)
	if err != nil {
		return
	}

	for rows.Next() {
		tag := Tag{}
		err = rows.Scan(&tag.Id, &tag.CreatedAt, &tag.Tag)
		if err != nil {
			return
		}

		tags = append(tags, tag)
	}

	return
}

func (app *App) getTagById(id string) (tag Tag, err error) {
	err = app.db.
		QueryRow("SELECT id, created_at, tag FROM tags WHERE id = ?", id).
		Scan(&tag.Id, &tag.CreatedAt, &tag.Tag)
	return
}

func (app *App) getMemesByTag(tag string) (memes []Meme, err error) {
	query := `
SELECT memes.id, memes.salt, memes.phash, memes.created_at, memes.path
FROM memes, tags
INNER JOIN memes_tags ON memes.id = memes_tags.meme_id AND tags.id = memes_tags.tag_id
WHERE tags.tag = ?`

	rows, err := app.db.Query(query, tag)
	if err != nil {
		return
	}

	for rows.Next() {
		meme := Meme{}
		err = rows.Scan(&meme.Id, &meme.Salt, &meme.PHash, &meme.CreatedAt, &meme.Filename)
		if err != nil {
			return
		}

		memes = append(memes, meme)
	}

	return
}

type MemeIndexElem struct {
	CreatedAt time.Time
	Filetype  string
	Tags      []string
}

func (app *App) reindexMeme(id string) (err error) {
	meme, tags, err := app.getMemeByIdWithTags(id)
	if err != nil {
		return
	}

	tagStringSlice := []string{}
	for _, tag := range tags {
		tagStringSlice = append(tagStringSlice, tag.Tag)
	}

	err = app.idx.Index(fmt.Sprintf("%v", meme.Id), MemeIndexElem{
		CreatedAt: meme.CreatedAt,
		Filetype:  path.Ext(meme.Filename),
		Tags:      tagStringSlice,
	})

	return
}

func (app *App) reindexMemes() (err error) {
	rows, err := app.db.Query("SELECT id FROM memes")
	if err != nil {
		return
	}

	for rows.Next() {
		id := ""
		err = rows.Scan(&id)
		if err != nil {
			return
		}

		err = app.reindexMeme(id)
		if err != nil {
			return
		}
	}

	return
}
