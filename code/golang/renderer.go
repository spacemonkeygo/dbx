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
	misc       *template.Template
	cre        *template.Template
	get        *template.Template
	get_all    *template.Template
	get_paged  *template.Template
	get_has    *template.Template
	get_count  *template.Template
	upd        *template.Template
	del        *template.Template
	del_all    *template.Template
	del_world  *template.Template
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

	r.misc, err = loader.Load("golang.misc.tmpl", nil)
	if err != nil {
		return nil, err
	}

	funcs := template.FuncMap{
		"sliceof":    sliceofFn,
		"param":      paramFn,
		"arg":        argFn,
		"zero":       zeroFn,
		"init":       initFn,
		"initnew":    initnewFn,
		"addrof":     addrofFn,
		"flatten":    flattenFn,
		"fieldvalue": fieldvalueFn,
		"comma":      commaFn,
	}

	r.cre, err = loader.Load("golang.create.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get, err = loader.Load("golang.get.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_all, err = loader.Load("golang.get-all.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_paged, err = loader.Load("golang.get-paged.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_has, err = loader.Load("golang.get-has.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_count, err = loader.Load("golang.get-count.tmpl", funcs)
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

	r.del_world, err = loader.Load("golang.delete-world.tmpl", funcs)
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

	// Render any result structs for multi-field reads
	for _, read := range root.Reads {
		if len(read.Selectables) < 2 {
			continue
		}
		if read.View == ir.Count || read.View == ir.Has {
			continue
		}
		s := ResultStructFromRead(read)
		if err := r.renderStruct(&buf, s); err != nil {
			return nil, err
		}
	}

	for _, dialect := range dialects {
		var gets []*ir.Model
		for _, cre := range root.Creates {
			gets = append(gets, cre.Model)
			if err := r.renderCreate(&buf, cre, dialect); err != nil {
				return nil, err
			}
		}
		for _, read := range root.Reads {
			if err := r.renderRead(&buf, read, dialect); err != nil {
				return nil, err
			}
		}
		for _, upd := range root.Updates {
			gets = append(gets, upd.Model)
			if err := r.renderUpdate(&buf, upd, dialect); err != nil {
				return nil, err
			}
		}
		for _, del := range root.Deletes {
			if err := r.renderDelete(&buf, del, dialect); err != nil {
				return nil, err
			}
		}

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

		if err := r.renderDeleteWorld(&buf, root.Models, dialect); err != nil {
			return nil, err
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
		Name       string
		SchemaSQL  string
		ExecOnOpen []string
	}

	type headerParams struct {
		Package      string
		ExtraImports []headerImport
		Dialects     []headerDialect
		Structs      []*ModelStruct
	}

	params := headerParams{
		Package: r.options.Package,
		Structs: ModelStructsFromIR(root.Models),
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
			Name:       dialect.Name(),
			SchemaSQL:  dialect_schema,
			ExecOnOpen: dialect.ExecOnOpen(),
		})
	}

	return tmplutil.Render(r.header, w, "", params)
}

func (r *Renderer) renderCreate(w io.Writer, ir_cre *ir.Create,
	dialect sql.Dialect) (err error) {

	cre := CreateFromIR(ir_cre, dialect)
	return r.renderFunc(r.cre, w, cre, dialect)
}

func (r *Renderer) renderRead(w io.Writer, ir_read *ir.Read,
	dialect sql.Dialect) error {

	get := GetFromIR(ir_read, dialect)
	switch ir_read.View {
	case ir.All, ir.LimitOffset:
		if ir_read.One() {
			if err := r.renderFunc(r.get, w, get, dialect); err != nil {
				return err
			}
		} else {
			if err := r.renderFunc(r.get_all, w, get, dialect); err != nil {
				return err
			}
		}
	case ir.Paged:
		if err := r.renderFunc(r.get_paged, w, get, dialect); err != nil {
			return err
		}
	case ir.Count:
		if err := r.renderFunc(r.get_count, w, get, dialect); err != nil {
			return err
		}
	case ir.Has:
		if err := r.renderFunc(r.get_has, w, get, dialect); err != nil {
			return err
		}
	default:
		return Error.New("unhandled read view %s", ir_read.View)
	}

	return nil
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

func (r *Renderer) renderDeleteWorld(w io.Writer, ir_models []*ir.Model,
	dialect sql.Dialect) error {

	type deleteWorld struct {
		SQLs []string
	}

	var del deleteWorld
	for i := len(ir_models) - 1; i >= 0; i-- {
		del.SQLs = append(del.SQLs, sql.RenderDelete(dialect, &ir.Delete{
			Model: ir_models[i],
		}))
	}

	return r.renderFunc(r.del_world, w, del, dialect)
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

	err = tmplutil.Render(r.misc, w, "decl", decl)
	if err != nil {
		return err
	}

	if isExported(decl.Signature) {
		r.signatures[decl.Signature] = true
	}

	return nil
}

func (r *Renderer) renderStruct(w io.Writer, s *Struct) (err error) {
	return tmplutil.Render(r.misc, w, "struct", s)
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
