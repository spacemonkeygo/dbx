// Copyright (C) 2017 Space Monkey, Inc.
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
	"bytes"
	"io"
	"text/template"
)

func RenderString(tmpl *template.Template, name string,
	data interface{}) (string, error) {

	var buf bytes.Buffer
	if err := Render(tmpl, &buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func Render(tmpl *template.Template, w io.Writer, name string,
	data interface{}) error {

	if name == "" {
		return Error.Wrap(tmpl.Execute(w, data))
	}
	return Error.Wrap(tmpl.ExecuteTemplate(w, name, data))
}
