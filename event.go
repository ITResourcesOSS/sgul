// Copyright 2019 Luca Stasio <joshuagame@gmail.com>
// Copyright 2019 IT Resources s.r.l.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sgul defines common structures and functionalities for applications.
// event.go defines commons for application events: normally to push some event
// into an amqp event bus.
package sgul

// Event is the struct used to push event messages into AMQP queues.
type Event struct {
	// Name is the global identifier for event. It MUST be
	// composed as "<action>_<resource>", a.e. "new_user", "upd_user", "del_user", ...
	Name string

	// Source is the global identifier for the client which push
	// the evt message. A.E. "uaa-servce", "acct-service", ...
	Source string

	// Payload is the struct containing all the event message information.
	// The AMQP Publisher will marshal it to json and the AMQP Subscriber
	// will unmarshal it into a specific request (something like a dto).
	Payload interface{}
}

// NewEvent return a new Event instance.
func NewEvent(name string, source string, payload interface{}) Event {
	return Event{
		Name:    name,
		Source:  source,
		Payload: payload,
	}
}
