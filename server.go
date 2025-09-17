package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/henriquebjr/shortener/url"
)

var (
	port    int
	baseUrl string
	stats   chan string
)

type Headers map[string]string

func init() {
	port = 8888
	baseUrl = fmt.Sprintf("http://localhost:%d", port)
}

func main() {
	stats = make(chan string)
	defer close(stats)
	go statsRegister(stats)

	url.Configure(url.NewRepository())

	http.HandleFunc("/api/shorten", Shortener)
	http.HandleFunc("/api/stats/", Viewer)
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
	answerWith(w, status, Headers{
		"Location": shortUrl,
		"Link": fmt.Sprintf("<%s/api/stats/%s>; rel=\"stats\"", baseUrl, url.Id),
	})
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

		stats <- id
	} else {
		http.NotFound(w, r)
	}
}

func Viewer(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	id := path[len(path)-1]

	if url := url.Find(id); url != nil {
		json, err := json.Marshal(url.Stats())

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		answerWithJson(w, string(json))
	} else {
		http.NotFound(w, r)
	}
}

func answerWithJson(w http.ResponseWriter, answer string) {
	answerWith(w, http.StatusOK, Headers{"Content-Type": "application/json"})
	fmt.Fprint(w, answer)
}

func statsRegister(ids <-chan string) {
	for id := range ids {
		url.Register(id)
		fmt.Printf("Redirect for %s.\n", id)
	}
}
