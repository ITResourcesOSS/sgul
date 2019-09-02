// Copyright 2019 Luca Stasio <joshuagame@gmail.com>
// Copyright 2019 IT Resources s.r.l.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// DoInTransaction (C) vertazzar on a comment at https://github.com/jinzhu/gorm/issues/2089
// Thanks for your useful tip.

// Package sgul defines common structures and functionalities for applications.
// controller.go defines commons for a base api controller.
package sgul

import (
	"net/http"

	"github.com/go-chi/chi"
)

// ClientError is an error whose details to be shared with client.
type ClientError interface {
	Error() string
	// ResponseBody returns response body.
	ResponseBody() ([]byte, error)
	// ResponseHeaders returns http status code and headers.
	ResponseHeaders() (int, map[string]string)
}

// HTTPError implements ClientError interface.
type HTTPError struct {
	Status  int         `json:"status"`
	Err     string      `json:"error"`
	Message interface{} `json:"detail"`
}

// ChiController defines the interface for an API Controller with Chi Router
type ChiController interface {
	Router() chi.Router
}

// Controller defines the base API Controller structure
type Controller struct {
	// Path is the base routing path for each route of the controller
	Path string
}

// RenderError returns error to the client
func (c *Controller) RenderError(w http.ResponseWriter, err error) {
	clientError, ok := err.(ClientError)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := clientError.ResponseBody()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	status, headers := clientError.ResponseHeaders()
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	w.Write(body)
}
