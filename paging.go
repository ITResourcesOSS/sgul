package sgul

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// Page defines the struct with paging info to send into the request context.
type Page struct {
	Page int
	Size int
}

type ctxPKey int

const ctxPageKey ctxPKey = iota

// ErrPagerNotInContext is returned if there is no Pager in the request context.
var ErrPagerNotInContext = errors.New("Pager info not in Context")

// Pager is the query paging middleware
func Pager() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			page := r.URL.Query().Get("page")
			size := r.URL.Query().Get("size")
			if page == "" {
				fmt.Println("page empty")
			} else {
				fmt.Printf("page: %s", page)
			}
			if size == "" {
				fmt.Println("page empty")
			} else {
				fmt.Printf("size: %s", size)
			}
		}
		return http.HandlerFunc(fn)
	}
}

// GetPage return the pager struct from request Context.
func GetPage(ctx context.Context) (Page, error) {
	if pager, ok := ctx.Value(ctxPageKey).(Page); ok {
		return pager, nil
	}
	return Page{}, ErrPagerNotInContext
}
