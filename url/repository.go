package url

type repository struct {
	urls map[string]*Url
}

func NewRepository() *repository {
	return &repository{make(map[string]*Url)}
}

func (r *repository) Exists(id string) bool {
	_, exists := r.urls[id]
	return exists
}

func (r *repository) Find(id string) *Url {
	return r.urls[id]
}

func (r *repository) FindByUrl(url string) *Url {
	for _, u  := range r.urls {
		if u.Destiny == url {
			return u
		}
	}
	return nil
}

func (r *repository) Save(url Url) error {
	r.urls[url.Id] = &url
	return nil
}