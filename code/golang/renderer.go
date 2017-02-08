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
	"fmt"
	"go/format"
	"io"
	"regexp"
	"sort"
	"text/template"
	"unicode"
	"unicode/utf8"

	"gopkg.in/spacemonkeygo/dbx.v1/code"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
	"gopkg.in/spacemonkeygo/dbx.v1/tmplutil"
)

var (
	reCollapseSpace = regexp.MustCompile(`\s+`)
)

type publicMethod struct {
	Signature string
	Invoke    string
}

type Options struct {
	Package   string
	SupportRx bool
}

type Renderer struct {
	loader          tmplutil.Loader
	header          *template.Template
	footer          *template.Template
	misc            *template.Template
	cre             *template.Template
	get_all         *template.Template
	get_has         *template.Template
	get_count       *template.Template
	get_limitoffset *template.Template
	get_paged       *template.Template
	get_scalar      *template.Template
	get_scalar_all  *template.Template
	get_one         *template.Template
	get_one_all     *template.Template
	get_first       *template.Template
	upd             *template.Template
	del             *template.Template
	del_all         *template.Template
	del_world       *template.Template
	get_last        *template.Template
	methods         map[string]publicMethod
	options         Options
}

var _ code.Renderer = (*Renderer)(nil)

func New(loader tmplutil.Loader, options *Options) (
	r *Renderer, err error) {

	r = &Renderer{
		loader:  loader,
		options: *options,
		methods: map[string]publicMethod{},
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
		"ctxparam":   ctxparamFn,
		"ctxarg":     ctxargFn,
	}

	r.cre, err = loader.Load("golang.create.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_all, err = loader.Load("golang.get-all.tmpl", funcs)
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

	r.get_paged, err = loader.Load("golang.get-paged.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_limitoffset, err = loader.Load("golang.get-limitoffset.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_scalar, err = loader.Load("golang.get-scalar.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_scalar_all, err = loader.Load("golang.get-scalar-all.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_one, err = loader.Load("golang.get-one.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_one_all, err = loader.Load("golang.get-one-all.tmpl", funcs)
	if err != nil {
		return nil, err
	}

	r.get_first, err = loader.Load("golang.get-first.tmpl", funcs)
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
	result_structs := map[string]bool{}
	for _, read := range root.Reads {
		if read.View == ir.Count || read.View == ir.Has {
			continue
		}
		if model := read.SelectedModel(); model != nil {
			continue
		}
		s := ResultStructFromRead(read)
		if result_structs[s.Name] {
			continue
		}
		result_structs[s.Name] = true
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

		if err = r.renderDialectFuncs(&buf, dialect); err != nil {
			return nil, err
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

	type headerDialect struct {
		Name       string
		SchemaSQL  string
		ExecOnOpen []string
	}

	type headerParams struct {
		Package      string
		ExtraImports []string
		Dialects     []headerDialect
		Structs      []*ModelStruct
		SupportRx    bool
	}

	params := headerParams{
		Package:   r.options.Package,
		Structs:   ModelStructsFromIR(root.Models),
		SupportRx: r.options.SupportRx,
	}

	for _, dialect := range dialects {
		dialect_schema := sql.RenderSchema(dialect, root)

		dialect_tmpl, err := r.loadDialect(dialect)
		if err != nil {
			return err
		}

		dialect_import, err := tmplutil.RenderString(dialect_tmpl, "import",
			nil)
		if err != nil {
			return err
		}

		params.ExtraImports = append(params.ExtraImports, dialect_import)
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

	var tmpl *template.Template
	switch ir_read.View {
	case ir.All:
		tmpl = r.get_all
	case ir.LimitOffset:
		tmpl = r.get_limitoffset
	case ir.Paged:
		tmpl = r.get_paged
	case ir.Count:
		tmpl = r.get_count
	case ir.Has:
		tmpl = r.get_has
	case ir.Scalar:
		if ir_read.Distinct() {
			tmpl = r.get_scalar
		} else {
			tmpl = r.get_scalar_all
		}
	case ir.One:
		if ir_read.Distinct() {
			tmpl = r.get_one
		} else {
			tmpl = r.get_one_all
		}
	case ir.First:
		tmpl = r.get_first
	default:
		panic(fmt.Sprintf("unhandled read view %s", ir_read.View))
	}

	return r.renderFunc(tmpl, w, get, dialect)
}

func (r *Renderer) renderUpdate(w io.Writer, ir_upd *ir.Update,
	dialect sql.Dialect) error {

	upd := UpdateFromIR(ir_upd, dialect)
	return r.renderFunc(r.upd, w, upd, dialect)
}

func (r *Renderer) renderDelete(w io.Writer, ir_del *ir.Delete,
	dialect sql.Dialect) error {

	del := DeleteFromIR(ir_del, dialect)
	if ir_del.Distinct() {
		return r.renderFunc(r.del, w, del, dialect)
	} else {
		return r.renderFunc(r.del_all, w, del, dialect)
	}
}

func (r *Renderer) renderDeleteWorld(w io.Writer, ir_models []*ir.Model,
	dialect sql.Dialect) error {

	type deleteWorld struct {
		Dialect string
		SQLs    []string
	}

	del := deleteWorld{
		Dialect: dialect.Name(),
	}
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

	method := publicMethod{
		Signature: signature.String(),
	}

	if isExported(method.Signature) {
		var invoke bytes.Buffer
		err = tmplutil.Render(tmpl, &invoke, "invoke", data)
		if err != nil {
			return err
		}
		method.Invoke = invoke.String()
		r.methods[method.Signature] = method
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
		Signature:    method.Signature,
		Body:         body.String(),
	}

	err = tmplutil.Render(r.misc, w, "decl", decl)
	if err != nil {
		return err
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
	var keys []string
	for key := range r.methods {
		keys = append(keys, key)
	}
	sort.Sort(sort.StringSlice(keys))

	type footerData struct {
		SupportRx bool
		Methods   []publicMethod
	}

	data := footerData{
		SupportRx: r.options.SupportRx,
	}

	for _, key := range keys {
		data.Methods = append(data.Methods, r.methods[key])
	}

	return tmplutil.Render(r.footer, w, "", data)
}

func (r *Renderer) renderDialectFuncs(w io.Writer, dialect sql.Dialect) (
	err error) {

	type dialectFunc struct {
		Receiver string
	}

	dialect_func := dialectFunc{
		Receiver: fmt.Sprintf("%sImpl", dialect.Name()),
	}

	tmpl, err := r.loadDialect(dialect)
	if err != nil {
		return err
	}

	return tmplutil.Render(tmpl, w, "is-constraint-error", dialect_func)
}

func (r *Renderer) loadDialect(dialect sql.Dialect) (
	*template.Template, error) {

	return r.loader.Load(
		fmt.Sprintf("golang.dialect-%s.tmpl", dialect.Name()), nil)
}
