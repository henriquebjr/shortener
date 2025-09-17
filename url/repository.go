package url

type repository struct {
	urls map[string]*Url
	counter map[string] int
}

func NewRepository() *repository {
	return &repository{
		make(map[string]*Url),
		make(map[string]int),
	}
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

func (r *repository) Register(id string) {
	r.counter[id]++
}

func (r *repository) RetrieveCounter(id string) int {
	return r.counter[id]
}