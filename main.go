package main

import (
	"html/template"
	"log"
	"net/http"
)

type PageData struct {
	RepoURL       string
	Branch        string
	Branches      []Branch
	DefaultBranch string
	RawURLs       []string
	Error         string
}

func main() {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{}

		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				data.Error = "Could not read the form data"
				renderTemplate(w, tmpl, data)
				return
			}
			repoURL := r.FormValue("repoUrl")
			selectedBranch := r.FormValue("branch")

			data.RepoURL = repoURL

			owner, repo, err := parseRepoURL(repoURL)
			if err != nil {
				data.Error = err.Error()
				renderTemplate(w, tmpl, data)
				return
			}

			branches, defaultBranch, err := getBranches(owner, repo)
			if err != nil {
				data.Error = err.Error()
				renderTemplate(w, tmpl, data)
				return
			}
			data.Branches = branches
			data.DefaultBranch = defaultBranch

			// Wenn kein Branch gewählt wurde, Default
			if selectedBranch == "" {
				selectedBranch = defaultBranch
			}
			data.Branch = selectedBranch

			raws, err := ListRawURLsWithOwnerRepo(owner, repo, selectedBranch)
			if err != nil {
				data.Error = err.Error()
				renderTemplate(w, tmpl, data)
				return
			}
			data.RawURLs = raws

			renderTemplate(w, tmpl, data)
			return
		}

		// GET: nur Anzeige, ohne Scan
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
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
