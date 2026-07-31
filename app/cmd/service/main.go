package main

import (
	"log"
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/views/pages"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/dist"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if err := pages.Preview().Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	addr := ":8080"
	log.Printf("preview server listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
