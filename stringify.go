// Copyright 2019 Luca Stasio <joshuagame@gmail.com>
// Copyright 2019 IT Resources s.r.l.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// DoInTransaction (C) vertazzar on a comment at https://github.com/jinzhu/gorm/issues/2089
// Thanks for your useful tip.

// Package sgul defines common structures and functionalities for applications.
// stringify.go converts a struct to its string representation.
package sgul

import (
	"bytes"
	"fmt"

	"github.com/fatih/structs"
)

// Stringify converts a struct to its string representation.
func Stringify(strct interface{}) string {
	var buffer bytes.Buffer
	buffer.WriteString("[")
	s := structs.New(strct)
	for _, f := range s.Fields() {
		buffer.WriteString(fmt.Sprintf(" %+v: <%+v>;", f.Name(), f.Value()))
	}
	buffer.WriteString(" ]")

	return buffer.String()
}
