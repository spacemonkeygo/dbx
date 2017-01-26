// Copyright (C) 2016 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ir

import (
	"strings"

	"bitbucket.org/pkg/inflect"
)

func (root *Root) SetDefaults() {
	for _, model := range root.Models {
		model.SetDefaults()
	}
	for _, cre := range root.Creates {
		cre.SetDefaults()
	}
	for _, read := range root.Reads {
		read.SetDefaults()
	}
	for _, upd := range root.Updates {
		upd.SetDefaults()
	}
	for _, del := range root.Deletes {
		del.SetDefaults()
	}
}

func (model *Model) SetDefaults() {
	if model.Table == "" {
		model.Table = inflect.Pluralize(model.Name)
	}

	for _, field := range model.Fields {
		field.SetDefaults()
	}

	for _, index := range model.Indexes {
		index.SetDefaults()
	}
}

func (field *Field) SetDefaults() {
	if field.Column == "" {
		field.Column = field.Name
	}
}

func (index *Index) SetDefaults() {
	if index.Name == "" {
		index.Name = DefaultIndexName(index)
	}
}

func (cre *Create) SetDefaults() {
	if cre.Suffix == "" {
		cre.Suffix = DefaultCreateSuffix(cre)
	}
}

func (read *Read) SetDefaults() {
	if read.Suffix == "" {
		read.Suffix = DefaultReadSuffix(read)
	}
}

func (upd *Update) SetDefaults() {
	if upd.Suffix == "" {
		upd.Suffix = DefaultUpdateSuffix(upd)
	}
}

func (del *Delete) SetDefaults() {
	if del.Suffix == "" {
		del.Suffix = DefaultDeleteSuffix(del)
	}
}

func DefaultIndexName(i *Index) string {
	parts := []string{i.Model.Table}
	for _, field := range i.Fields {
		parts = append(parts, field.Column)
	}
	if i.Unique {
		parts = append(parts, "unique")
	}
	parts = append(parts, "index")
	return strings.Join(parts, "_")
}

func DefaultCreateSuffix(cre *Create) string {
	var parts []string
	if cre.Raw {
		parts = append(parts, "raw")
	}
	parts = append(parts, cre.Model.Name)
	return strings.Join(parts, "_")
}

func DefaultReadSuffix(read *Read) string {
	var parts []string
	for _, selectable := range read.Selectables {
		part := selectable.UnderRef()
		if !read.One() {
			part = inflect.Pluralize(part)
		}
		parts = append(parts, part)
	}
	parts = append(parts, whereSuffix(read.Where, len(read.Joins) > 0)...)
	switch read.View {
	case All, Count, Has:
	case LimitOffset:
		parts = append(parts, "with", "limit", "offset")
	case Paged:
		parts = append(parts, "paged")
	}
	return strings.Join(parts, "_")
}

func DefaultUpdateSuffix(upd *Update) string {
	var parts []string
	parts = append(parts, upd.Model.Name)
	parts = append(parts, whereSuffix(upd.Where, len(upd.Joins) > 0)...)
	return strings.Join(parts, "_")
}

func DefaultDeleteSuffix(del *Delete) string {
	var parts []string
	parts = append(parts, del.Model.Name)
	parts = append(parts, whereSuffix(del.Where, len(del.Joins) > 0)...)
	return strings.Join(parts, "_")
}

func whereSuffix(wheres []*Where, full bool) (parts []string) {
	if len(wheres) == 0 {
		return nil
	}
	parts = append(parts, "by")
	for i, where := range wheres {
		if where.Right != nil {
			continue
		}
		if i > 0 {
			parts = append(parts, "and")
		}
		if full {
			parts = append(parts, where.Left.Model.Name)
		}
		parts = append(parts, where.Left.Name)
		if suffix := where.Op.Suffix(); suffix != "" {
			parts = append(parts, suffix)
		}
	}
	return parts
}
