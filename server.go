package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/henriquebjr/shortener/url"
)

var (
	port    int
	baseUrl string
)

type Headers map[string]string

func init() {
	port = 8888
	baseUrl = fmt.Sprintf("http://localhost:%d", port)
}

func main() {
	url.Configure(url.NewRepository())

	http.HandleFunc("/api/shorten", Shortener)
	http.HandleFunc("/r/", Redirector)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func Shortener(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		answerWith(w, http.StatusMethodNotAllowed, Headers{
			"Allow": "POST",
		})
		return
	}

	url, isNew, err := url.FindOrCreateNewUrl(extractUrl(r))

	if err != nil {
		answerWith(w, http.StatusBadRequest, nil)
		return
	}

	var status int
	if isNew {
		status = http.StatusCreated
	} else {
		status = http.StatusOK
	}

	shortUrl := fmt.Sprintf("%s/r/%s", baseUrl, url.Id)
	answerWith(w, status, Headers{"Location": shortUrl})
}

func answerWith(w http.ResponseWriter, status int, headers Headers) {
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
}

func extractUrl(r *http.Request) string {
	url := make([]byte, r.ContentLength, r.ContentLength)
	r.Body.Read(url)
	return string(url)
}

func Redirector(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	id := path[len(path)-1]

	if url := url.Find(id); url != nil {
		http.Redirect(w, r, url.Destiny, http.StatusMovedPermanently)
	} else {
		http.NotFound(w, r)
	}
}