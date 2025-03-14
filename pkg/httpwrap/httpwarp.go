package httpwrap

import (
	"net/http"
	"slices"
)

func Methods(methods []string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !slices.Contains(methods, r.Method) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		handler(w, r)
	}
}

func GET(handler http.HandlerFunc) http.HandlerFunc {
	return Methods([]string{http.MethodGet}, handler)
}

func POST(handler http.HandlerFunc) http.HandlerFunc {
	return Methods([]string{http.MethodPost}, handler)
}
