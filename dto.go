// Copyright 2019 Luca Stasio <joshuagame@gmail.com>
// Copyright 2019 IT Resources s.r.l.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// DoInTransaction (C) vertazzar on a comment at https://github.com/jinzhu/gorm/issues/2089
// Thanks for your useful tip.

// Package sgul defines common structures and functionalities for applications.
// dto.go defines commons for a DTo object (just to have a String() func).
package sgul

import (
	"bytes"
	"fmt"

	"github.com/fatih/structs"
)

// DTO represent the base struct for a DTO.
// Defines the String() func to be used to log out struct values.
type DTO struct{}

func (dto *DTO) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("[")
	s := structs.New(dto)
	for _, f := range s.Fields() {
		buffer.WriteString(fmt.Sprintf(" %+v: <%+v>;", f.Name(), f.Value()))
	}
	buffer.WriteString(" ]")

	return buffer.String()
}
