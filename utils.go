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

// MaskedStringify converts a struct to its string representation obfuscating values
// for the key passed in mask slice.
func MaskedStringify(strct interface{}, mask []string) string {
	var buffer bytes.Buffer
	buffer.WriteString("[")
	s := structs.New(strct)
	for _, f := range s.Fields() {
		if ContainsString(mask, f.Name()) {
			buffer.WriteString(fmt.Sprintf(" %+v: <**********>;", f.Name()))
		} else {
			buffer.WriteString(fmt.Sprintf(" %+v: <%+v>;", f.Name(), f.Value()))
		}
	}
	buffer.WriteString(" ]")

	return buffer.String()
}

// ContainsString checks if a slice of strings contains a string.
func ContainsString(s []string, elem string) bool {
	for _, a := range s {
		if a == elem {
			return true
		}
	}
	return false
}

// MergeStringSlices merges string slices avoiding duplicates.
// source code from Jacy Gao (http://jgao.io/?p=119)... thank you man!
func MergeStringSlices(s1, s2 []string) []string {
	s1 = append(s1, s2...)
	seen := make(map[string]struct{}, len(s1))
	j := 0
	for _, s := range s1 {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		s1[j] = s
		j++
	}

	mergedSlice := s1[:j]
	return mergedSlice
}
