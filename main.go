package main

import (
	"html/template"
	"log"
	"net/http"
)

type PageData struct {
	RepoURL string
	Branch  string
	RawURLs []string
	Error   string
}

func main() {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{}

		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				data.Error = "Formular konnte nicht gelesen werden"
				renderTemplate(w, tmpl, data)
				return
			}
			repoURL := r.FormValue("repoUrl")
			branch := r.FormValue("branch")

			data.RepoURL = repoURL
			data.Branch = branch

			raws, err := ListRawURLs(repoURL, branch)
			if err != nil {
				data.Error = err.Error()
				renderTemplate(w, tmpl, data)
				return
			}
			data.RawURLs = raws
		}

		renderTemplate(w, tmpl, data)
	})

	addr := ":8080"
	log.Printf("Starte Server auf %s ...", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "index", data); err != nil {
		http.Error(w, "Templatefehler", http.StatusInternalServerError)
	}
}
