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
	"unicode"
	"unicode/utf8"

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
	funcs      *template.Template
	ins        *template.Template
	sel        *template.Template
	upd        *template.Template
	del        *template.Template
	del_all    *template.Template
	get_last   *template.Template
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

	r.funcs, err = loader.Load("golang.funcs.tmpl", nil)
	if err != nil {
		return nil, err
	}

	funcs := template.FuncMap{
		"param":    paramFn,
		"arg":      argFn,
		"zero":     zeroFn,
		"init":     initFn,
		"initnew":  initnewFn,
		"autoinit": autoinitFn,
		"addrof":   addrofFn,
		"flatten":  flattenFn,
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

	r.del_all, err = loader.Load("golang.delete-all.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_last, err = loader.Load("golang.get-last.tmpl", funcs)
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
		var gets []*ir.Model
		for _, ins := range root.Inserts {
			gets = append(gets, ins.Model)
			if err := r.renderInsert(&buf, ins, dialect); err != nil {
				return nil, err
			}
		}
		//		for _, sel := range root.Selects {
		//			if err := r.renderSelect(&buf, sel, dialect); err != nil {
		//				return nil, err
		//			}
		//		}
		//		for _, upd := range root.Updates {
		//			gets = append(gets, upd.Model)
		//			if err := r.renderUpdate(&buf, upd, dialect); err != nil {
		//				return nil, err
		//			}
		//		}
		for _, del := range root.Deletes {
			if err := r.renderDelete(&buf, del, dialect); err != nil {
				return nil, err
			}
		}
		// 	if err := r.renderDelete(&buf, del, dialect); err != nil {
		// 		return nil, err
		// 	}
		// }

		if len(gets) > 0 && !dialect.Features().Returning {
			// dialect does not support returning columns on insert and updates
			// so we need to generate a function to support getting by last
			// insert id.
			done := map[*ir.Model]bool{}
			for _, model := range gets {
				if done[model] {
					continue
				}
				done[model] = true
				if err := r.renderGetLast(&buf, model, dialect); err != nil {
					return nil, err
				}
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

func (r *Renderer) renderInsert(w io.Writer, ir_ins *ir.Insert,
	dialect sql.Dialect) (err error) {

	ins := InsertFromIR(ir_ins, dialect)
	return r.renderFunc(r.ins, w, ins, dialect)
}

func (r *Renderer) renderSelect(w io.Writer, ir_sel *ir.Select,
	dialect sql.Dialect) error {

	sel := SelectFromIR(ir_sel, dialect)
	return r.renderFunc(r.sel, w, sel, dialect)
}

func (r *Renderer) renderUpdate(w io.Writer, ir_upd *ir.Update,
	dialect sql.Dialect) error {

	upd := UpdateFromIR(ir_upd, dialect)
	return r.renderFunc(r.upd, w, upd, dialect)
}

func (r *Renderer) renderDelete(w io.Writer, ir_del *ir.Delete,
	dialect sql.Dialect) error {

	del := DeleteFromIR(ir_del, dialect)
	if ir_del.One() {
		return r.renderFunc(r.del, w, del, dialect)
	} else {
		return r.renderFunc(r.del_all, w, del, dialect)
	}
}

func (r *Renderer) renderFunc(tmpl *template.Template, w io.Writer,
	data interface{}, dialect sql.Dialect) (err error) {

	var signature bytes.Buffer
	err = tmplutil.Render(tmpl, &signature, "signature", data)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	err = tmplutil.Render(tmpl, &body, "body", data)
	if err != nil {
		return err
	}

	type funcDecl struct {
		ReceiverBase string
		Signature    string
		Body         string
	}

	decl := funcDecl{
		ReceiverBase: dialect.Name(),
		Signature:    signature.String(),
		Body:         body.String(),
	}

	err = tmplutil.Render(r.funcs, w, "decl", decl)
	if err != nil {
		return err
	}

	if isExported(decl.Signature) {
		r.signatures[decl.Signature] = true
	}

	return nil
}

func isExported(signature string) bool {
	r, _ := utf8.DecodeRuneInString(signature)
	return unicode.IsUpper(r)
}

func (r *Renderer) renderGetLast(w io.Writer, model *ir.Model,
	dialect sql.Dialect) error {

	type getLast struct {
		Return *Var
		SQL    string
	}

	get_last := getLast{
		Return: VarFromModel(model),
		SQL:    sql.RenderGetLast(dialect, model),
	}

	return r.renderFunc(r.get_last, w, get_last, dialect)
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
