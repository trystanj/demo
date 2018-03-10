package main

var memStore = []Result{
	Result{
		ID:       "0",
		Name:     "a",
		Category: "bar",
	},
	Result{
		ID:       "1",
		Name:     "b",
		Category: "restaurant",
	},
	Result{
		ID:       "2",
		Name:     "c",
		Category: "clerb",
	},
	Result{
		ID:       "3",
		Name:     "d",
		Category: "bar",
	},
}

type MemStore struct {
	host  string
	store []Result
}

func NewMemStore(host string) *MemStore {
	return &MemStore{
		host:  host,
		store: memStore,
	}
}

func (m *MemStore) Fetch(from int, to int, category string) (*Results, error) {
	filtered := make([]Result, 0)

	for _, v := range memStore {
		if v.Category == category {
			filtered = append(filtered, v)
		}
	}

	// hack
	if from > len(filtered) {
		from = len(filtered) - 1
	}
	if to > len(filtered) {
		to = len(filtered)
	}

	results := &Results{
		Results: filtered[from:to],
		Token:   to,
	}

	return results, nil
}
