package url

import (
	"math/rand"
	"net/url"
	"time"
)

const (
	size    = 5
	symbols = "abcdefghijklmnopqrstuvxyzABCDEFGHIJKLMNOPQRSTUVXYZ1234567890_-+"
)

type Repository interface {
	Exists(id string) bool
	Find(id string) *Url
	FindByUrl(url string) *Url
	Save(url Url) error
}

type Url struct {
	Id       string
	Creation time.Time
	Destiny  string
}

var repo Repository

func Configure(r Repository) {
	repo = r
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Find(id string) *Url {
	return repo.Find(id)
}

func FindOrCreateNewUrl(destiny string) (u *Url, isNew bool, err error) {
	if u = repo.FindByUrl(destiny); u != nil {
		return u, false, nil
	}

	if _, err = url.ParseRequestURI(destiny); err != nil {
		return nil, false, err
	}

	url := Url{generateId(), time.Now(), destiny}
	repo.Save(url)
	return &url, true, nil
}

func generateId() string {
	newId := func() string {
		id := make([]byte, size, size)
		for i := range id {
			id[i] = symbols[rand.Intn(len(symbols))]
		}
		return string(id)
	}

	for {
		if id := newId(); !repo.Exists(id) {
			return id
		}
	}
}
