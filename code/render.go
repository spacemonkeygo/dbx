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

package code

import (
	"io"

	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
)

func Render(w io.Writer, root *ir.Root, language Language,
	dialects []sql.Dialect) (err error) {

	r := renderer{w: w, language: language}
	return r.render(root, dialects)
}

type renderer struct {
	w        io.Writer
	language Language
	err      error
}

func (r *renderer) render(root *ir.Root, dialects []sql.Dialect) (err error) {
	r.setError(r.language.RenderHeader(r.w, root, dialects))
	for _, dialect := range dialects {
		r.renderDialect(root, dialect)
	}
	r.setError(r.language.RenderFooter(r.w, root, dialects))

	return r.err
}

func (r *renderer) renderDialect(root *ir.Root, dialect sql.Dialect) {
	if !r.ok() {
		return
	}

	for _, model := range root.Models.Models() {
		r.renderInsert(model, dialect)
	}
	for _, sel := range root.Selects {
		r.renderSelect(sel, dialect)
	}
	for _, del := range root.Deletes {
		r.renderDelete(del, dialect)
	}
}

func (r *renderer) renderInsert(model *ir.Model, dialect sql.Dialect) {
	if !r.ok() {
		return
	}
	r.setError(r.language.RenderInsert(r.w, model, dialect))
}

func (r *renderer) renderSelect(sel *ir.Select, dialect sql.Dialect) {
	if !r.ok() {
		return
	}
	r.setError(r.language.RenderSelect(r.w, sel, dialect))
}

func (r *renderer) renderDelete(del *ir.Delete, dialect sql.Dialect) {
	if !r.ok() {
		return
	}
	r.setError(r.language.RenderDelete(r.w, del, dialect))
}

func (r *renderer) ok() bool {
	return r.err == nil
}

func (r *renderer) setError(err error) (ok bool) {
	if err != nil && r.err == nil {
		r.err = err
	}
	return r.err == nil
}
