package sgul

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/middleware"
)

// Page defines the struct with paging info to send into the request context.
type Page struct {
	Page int
	Size int
}

type ctxPKey int

const ctxPageKey ctxPKey = iota + 1

// ErrPagerNotInContext is returned if there is no Pager in the request context.
var ErrPagerNotInContext = errors.New("Pager info not in Context")

func pager(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get("page")
		s := r.URL.Query().Get("size")
		if p != "" && s != "" {
			var pVal int
			var sVal int
			var err error
			pVal, err = strconv.Atoi(p)
			if err != nil {
				RenderError(w, NewHTTPError(err, http.StatusBadRequest, "Malformed 'page' param", middleware.GetReqID(r.Context())))
				return
			}
			sVal, err = strconv.Atoi(s)
			if err != nil {
				RenderError(w, NewHTTPError(err, http.StatusBadRequest, "Malformed 'size' param", middleware.GetReqID(r.Context())))
				return
			}
			page := Page{Page: pVal, Size: sVal}
			ctx := context.WithValue(r.Context(), ctxPageKey, page)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			next.ServeHTTP(w, r)
		}
	}
}

// Pager is the query paging middleware.
func Pager() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(pager(next))
	}
}

// RoutePager is the query paging middleware to be used on routes.
func RoutePager(next http.HandlerFunc) http.HandlerFunc {
	return pager(next)
}

// GetPage return the pager struct from request Context.
func GetPage(ctx context.Context) (Page, error) {
	if page, ok := ctx.Value(ctxPageKey).(Page); ok {
		return page, nil
	}

	return Page{}, ErrPagerNotInContext
}
