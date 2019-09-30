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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

type (
	// ClientError is an error whose details to be shared with client.
	ClientError interface {
		Error() string
		// ResponseBody returns response body.
		ResponseBody() ([]byte, error)
		// ResponseHeaders returns http status code and headers.
		ResponseHeaders() (int, map[string]string)
	}

	// HTTPError implements ClientError interface.
	HTTPError struct {
		Code      int         `json:"code"`
		Err       string      `json:"error"`
		Detail    interface{} `json:"detail"`
		RequestID string      `json:"requestId"`
		Timestamp time.Time   `json:"timestamp"`
	}

	// ChiController defines the interface for an API Controller with Chi Router
	ChiController interface {
		Router() chi.Router
	}

	// RestController is the Rest API Controller interface.
	RestController interface {
		BasePath() string
		Router() chi.Router
	}

	// Controller defines the base API Controller structure
	Controller struct {
		// Path is the base routing path for each route of the controller
		Path string
	}
)

// HTTPError ------------------------------------------------------------

// Error return a formatted description of the error.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("%v : %s", e.Detail, e.Err)
}

// ResponseBody returns JSON response body.
func (e *HTTPError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing response body: %v", err)
	}
	return body, nil
}

// ResponseHeaders returns http status code and headers.
func (e *HTTPError) ResponseHeaders() (int, map[string]string) {
	return e.Code, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
}

// NewHTTPError returns a new HTTPError instance
func NewHTTPError(err error, status int, detail interface{}, requestID string) error {

	return &HTTPError{
		Err:       err.Error(),
		Detail:    detail,
		Code:      status,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
}

// Controller ------------------------------------------------------------

// NewController return a new instance of Controller (useful in composition for api controllers).
func NewController(path string) Controller {
	return Controller{Path: path}
}

// RenderError returns error to the client.
func (c *Controller) RenderError(w http.ResponseWriter, err error) {
	RenderError(w, err)
}

// RenderError exported to be used in this lib (a.e. in middlewares) returns error to the client.
func RenderError(w http.ResponseWriter, err error) {
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
