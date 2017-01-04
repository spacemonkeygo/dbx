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

package tmplutil

import (
	"io/ioutil"
	"path/filepath"
	"text/template"

	"bitbucket.org/pkg/inflect"
)

type Loader interface {
	Load(name string, funcs template.FuncMap) (*template.Template, error)
}

type LoaderFunc func(name string, funcs template.FuncMap) (
	*template.Template, error)

func (fn LoaderFunc) Load(name string, funcs template.FuncMap) (
	*template.Template, error) {

	return fn(name, funcs)
}

type DirLoader string

func (d DirLoader) Load(name string, funcs template.FuncMap) (
	*template.Template, error) {

	data, err := ioutil.ReadFile(filepath.Join(string(d), name))
	if err != nil {
		return nil, Error.Wrap(err)
	}
	return loadTemplate(name, data, funcs)
}

type BinLoader func(name string) ([]byte, error)

func (b BinLoader) Load(name string, funcs template.FuncMap) (
	*template.Template, error) {

	data, err := b(name)
	if err != nil {
		return nil, Error.Wrap(err)
	}
	return loadTemplate(name, data, funcs)
}

func loadTemplate(name string, data []byte, funcs template.FuncMap) (
	*template.Template, error) {

	if funcs == nil {
		funcs = make(template.FuncMap)
	}

	safeset := func(name string, fn interface{}) {
		if funcs[name] == nil {
			funcs[name] = fn
		}
	}

	safeset("pluralize", inflect.Pluralize)
	safeset("singularize", inflect.Singularize)
	safeset("camelize", inflect.Camelize)
	safeset("cameldown", inflect.CamelizeDownFirst)
	safeset("underscore", inflect.Underscore)

	tmpl, err := template.New(name).Funcs(funcs).Parse(string(data))
	return tmpl, Error.Wrap(err)
}
