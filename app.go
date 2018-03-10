package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const (
	pageSize = 20
)

// Result is one location
type Result struct {
	ID       string
	Name     string
	Category string
}

// Results is the api output
type Results struct {
	Results []Result
	Token   int
}

// Fetcher pulls results from a store
type Fetcher interface {
	Fetch(from int, to int, category string) (*Results, error)
}

type app struct {
	fetcher Fetcher
}

// search returns Results given a pagination token
func (app *app) search() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var paginationToken int
		var category string
		var err error

		// pagination token - initial request does not need to start with a page, so set to 0 if not present
		token := r.URL.Query().Get("token")
		if token == "" {
			paginationToken = 0
		} else {
			paginationToken, err = strconv.Atoi(token)
			if err != nil {
				message := "invalid request"
				http.Error(w, message, http.StatusBadRequest)
				return
			}
		}

		categoryParam := r.URL.Query().Get("category")

		// set default category if none present
		if categoryParam == "" {
			category = "restaurant"
		} else {
			category = categoryParam
		}

		results, err := app.fetcher.Fetch(paginationToken, paginationToken+pageSize, category)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "oh shit")
		}

		err = json.NewEncoder(w).Encode(results)
		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, http.StatusText(status), status)
			return
		}
	})
}

/////////////////////////
//   Internal Routes   //
/////////////////////////

func (app *app) hint() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		message := `Hit /search to begin a search. You can pass in the following tokens:
		token: Used as a pagination token to continue searching from last results (optional)
		category: Which category to search in [restaurant, bar, clerb]
		`

		fmt.Fprintf(w, message)
		return
	})
}
