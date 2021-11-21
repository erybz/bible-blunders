package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	db   DB
	tmpl *template.Template
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(h.db)
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	h.tmpl.Execute(w, h.db)
}

func main() {
	db, err := LoadDB("db.json")
	if err != nil {
		log.Fatalf("failed to load database file:%v", err)
	}

	tmpl := template.Must(
		template.New("index.gohtml").Funcs(sprig.FuncMap()).ParseGlob("*.gohtml"),
	)

	h := Handler{db, tmpl}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/db", h.Handle)
	r.Get("/", h.Index)

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "public"))
	FileServer(r, "/static", filesDir)

	http.ListenAndServe(":8000", r)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func LoadDB(name string) (DB, error) {
	file, err := os.Open(name)
	if err != nil {
		return DB{}, err
	}
	defer file.Close()

	var d DB
	if err := json.NewDecoder(file).Decode(&d); err != nil {
		return DB{}, err
	}
	return d, nil
}

type DB struct {
	Index   []string  `json:"index"`
	Content []Content `json:"content"`
}
type Content struct {
	Title  string   `json:"title"`
	Verses []Verses `json:"verses"`
}

type Verses struct {
	Title     string   `json:"title"`
	Pre       []string `json:"pre"`
	Post      []string `json:"post"`
	Verse     string   `json:"verse"`
	Reference string   `json:"reference"`
}
