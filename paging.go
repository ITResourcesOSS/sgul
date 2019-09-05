package sgul

import (
	"context"
	"errors"
	"fmt"
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

// Pager is the query paging middleware
func Pager() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			p := r.URL.Query().Get("page")
			s := r.URL.Query().Get("size")
			if p != "" && s != "" {
				fmt.Printf("GOT PAGE INFO %s,%s\n", p, s)
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
				fmt.Printf("GOING WITH CTX PAGE INFO %d,%d\n", page.Page, page.Size)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				fmt.Println("GOING WITHOUT CTX PAGE INFO")
				next.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(fn)
	}
}

// GetPage return the pager struct from request Context.
func GetPage(ctx context.Context) (Page, error) {
	if page, ok := ctx.Value(ctxPageKey).(Page); ok {
		fmt.Println("*** Fund Page in Context")
		return page, nil
	}

	fmt.Println("*** No Page in context")
	return Page{}, ErrPagerNotInContext
}
