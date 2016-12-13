package main

import (
	"database/sql"
	"github.com/blevesearch/bleve"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"net/http"
	"os"
	"path"
)

type AppConfig struct {
	HttpListen  string `json:"http-listen"`
	BleveIndex  string `json:"bleve-index"`
	Database    string `json:"database"`
	TemplateDir string `json:"template-dir"`
	UploadsDir  string `json:"uploads-dir"`
	UploadsPath string `json:"uploads-path"`
	Secret      string `json:"secret"`
}

type App struct {
	cfg     AppConfig
	db      *sql.DB
	tmpl    *template.Template
	idx     bleve.Index
	handler http.Handler
}

func CreateApp(cfg AppConfig) (app *App) {
	app = &App{
		cfg: cfg,
	}

	// Open the database and run migrations
	db, err := sql.Open("sqlite3", app.cfg.Database)
	if err != nil {
		panic(err)
	}
	app.db = db
	app.Migrate()

	// Setup the bleve search index
	if _, err := os.Stat(app.cfg.BleveIndex); os.IsNotExist(err) {
		// Create a new index
		mapping := bleve.NewIndexMapping()
		index, err := bleve.New(app.cfg.BleveIndex, mapping)
		if err != nil {
			panic(err)
		}
		app.idx = index
	} else {
		// Open an existing index
		index, err := bleve.Open(app.cfg.BleveIndex)
		if err != nil {
			panic(err)
		}
		app.idx = index
	}

	// Precompile templates
	app.tmpl = template.Must(template.ParseGlob(path.Join(app.cfg.TemplateDir, "*.html")))

	// Setup routes
	r := mux.NewRouter()
	r.HandleFunc("/", app.HomeHandler)
	r.HandleFunc("/search", app.SearchHandler)
	r.HandleFunc("/upload", app.UploadHandler)
	r.HandleFunc("/tags", app.TagsHandler)
	r.HandleFunc("/duplicates/{id}", app.DuplicatesHandler)
	r.HandleFunc("/edit/{id}", app.EditHandler)
	r.HandleFunc("/delete/{id}", app.DeleteHandler)
	r.PathPrefix(app.cfg.UploadsPath).Handler(http.StripPrefix(app.cfg.UploadsPath, http.FileServer(http.Dir(app.cfg.UploadsDir))))

	// Apply middlewares
	app.handler = RecoverMiddleware(r)
	app.handler = csrf.Protect([]byte(app.cfg.Secret), csrf.Secure(false))(app.handler)

	return
}

func (app *App) ListenAndServe() {
	defer app.db.Close()
	panic(http.ListenAndServe(app.cfg.HttpListen, app.handler))
}

func (app *App) renderTemplate(w http.ResponseWriter, r *http.Request, templateFile string, localData map[string]interface{}) {
	// Global template data
	data := map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	}

	// Merge the local template data with the global one
	for k, v := range localData {
		data[k] = v
	}

	app.tmpl.ExecuteTemplate(w, templateFile, data)
}
