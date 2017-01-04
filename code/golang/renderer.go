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

package golang

import (
	"bytes"
	"go/format"
	"io"
	"regexp"
	"sort"
	"text/template"

	"github.com/spacemonkeygo/errors"
	"gopkg.in/spacemonkeygo/dbx.v1/code"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
	"gopkg.in/spacemonkeygo/dbx.v1/tmplutil"
)

var (
	Error = errors.NewClass("golang")

	reCollapseSpace = regexp.MustCompile(`\s+`)
)

type Options struct {
	Package string
}

type Renderer struct {
	header     *template.Template
	footer     *template.Template
	ins        *template.Template
	sel        *template.Template
	upd        *template.Template
	del        *template.Template
	signatures map[string]bool
	options    Options
}

var _ code.Renderer = (*Renderer)(nil)

func New(loader tmplutil.Loader, options *Options) (
	r *Renderer, err error) {

	r = &Renderer{
		options:    *options,
		signatures: map[string]bool{},
	}

	r.header, err = loader.Load("golang.header.tmpl", nil)
	if err != nil {
		return nil, err
	}

	r.footer, err = loader.Load("golang.footer.tmpl", nil)
	if err != nil {
		return nil, err
	}

	funcs := template.FuncMap{
		"params": asParam,
		"args":   asArg,
		"zeroed": asZero,
	}

	r.ins, err = loader.Load("golang.insert.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.sel, err = loader.Load("golang.select.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.upd, err = loader.Load("golang.update.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.del, err = loader.Load("golang.delete.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Renderer) RenderCode(root *ir.Root, dialects []sql.Dialect) (
	rendered []byte, err error) {
	var buf bytes.Buffer

	if err := r.renderHeader(&buf, root, dialects); err != nil {
		return nil, err
	}

	for _, dialect := range dialects {
		//		for _, ins := range root.Inserts {
		//			if err := r.renderInsert(&buf, ins, dialect); err != nil {
		//				return nil, err
		//			}
		//		}
		//		//		for _, sel := range root.Selects {
		//		//			if err := r.renderSelect(&buf, sel, dialect); err != nil {
		//		//				return nil, err
		//		//			}
		//		//		}
		//		//		for _, upd := range root.Updates {
		//		//			if err := r.renderUpdate(&buf, upd, dialect); err != nil {
		//		//				return nil, err
		//		//			}
		//		//		}
		for _, del := range root.Deletes {
			if err := r.renderDelete(&buf, del, dialect); err != nil {
				return nil, err
			}
		}
	}

	if err := r.renderFooter(&buf); err != nil {
		return nil, err
	}

	rendered, err = format.Source(buf.Bytes())
	if err != nil {
		rendered = buf.Bytes()
		//	return nil, Error.Wrap(err)
	}

	return rendered, nil
}

func (r *Renderer) renderHeader(w io.Writer, root *ir.Root,
	dialects []sql.Dialect) error {

	type headerImport struct {
		As      string
		Package string
	}

	type headerDialect struct {
		Name      string
		SchemaSQL string
	}

	type headerParams struct {
		Package        string
		ExtraImports   []headerImport
		Dialects       []headerDialect
		Structs        []*Struct
		StructsReverse []*Struct
	}

	params := headerParams{
		Package: r.options.Package,
		Structs: StructsFromIR(root.Models.Models()),
	}

	for i := len(params.Structs) - 1; i >= 0; i-- {
		params.StructsReverse = append(params.StructsReverse, params.Structs[i])
	}

	for _, dialect := range dialects {
		dialect_schema := sql.RenderSchema(dialect, root)

		var driver string
		switch dialect.Name() {
		case "postgres":
			driver = "github.com/lib/pq"
		case "sqlite3":
			driver = "github.com/mattn/go-sqlite3"
		default:
			return Error.New("unsupported dialect %q", dialect.Name())
		}

		params.ExtraImports = append(params.ExtraImports, headerImport{
			As:      "_",
			Package: driver,
		})

		params.Dialects = append(params.Dialects, headerDialect{
			Name:      dialect.Name(),
			SchemaSQL: dialect_schema,
		})
	}

	return tmplutil.Render(r.header, w, "", params)
}

func (r *Renderer) renderInsert(w io.Writer, ins *ir.Insert,
	dialect sql.Dialect) (err error) {

	go_ins := InsertFromIR(ins, dialect)
	if err := tmplutil.Render(r.ins, w, "", go_ins); err != nil {
		return err
	}

	return nil
}

func (r *Renderer) renderSelect(w io.Writer, sel *ir.Select,
	dialect sql.Dialect) error {

	go_sel := SelectFromIR(sel, dialect)
	if err := tmplutil.Render(r.sel, w, "", go_sel); err != nil {
		return err
	}

	return nil
}

func (r *Renderer) renderUpdate(w io.Writer, upd *ir.Update,
	dialect sql.Dialect) error {

	if err := tmplutil.Render(r.upd, w, "", nil); err != nil {
		return err
	}

	return nil
}

func (r *Renderer) renderDelete(w io.Writer, del *ir.Delete,
	dialect sql.Dialect) error {

	data := DeleteFromIR(del, dialect)

	return tmplutil.Render(r.del, w, "", data)
}

func (r *Renderer) renderFooter(w io.Writer) error {
	var funcs []string
	for k := range r.signatures {
		funcs = append(funcs, k)
	}
	sort.Sort(sort.StringSlice(funcs))

	return tmplutil.Render(r.footer, w, "", map[string]interface{}{
		"Funcs": funcs,
	})
}
