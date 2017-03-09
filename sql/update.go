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

package sql

import "gopkg.in/spacemonkeygo/dbx.v1/ir"

const (
	updateTmplPrefix = `
	UPDATE {{ .Table }} SET`
	updateTmplSuffix = `
	{{ if .Where }} WHERE {{- range $i, $w := .Where }}{{ if $i }} AND{{ end }} {{ $w.Left }} {{ $w.Op }} {{ $w.Right }}{{ end }} {{- end -}}
	{{- if .Returning }} RETURNING
	{{- range $i, $col := .Returning }}
		{{- if $i }},{{ end }} {{ $col }}
	{{- end }}
	{{- end }}`
)

func RenderUpdate(dialect Dialect, ir_upd *ir.Update) (prefix, suffix string) {
	upd := UpdateFromIR(ir_upd, dialect)
	prefix = render(dialect, updateTmplPrefix, upd, noTerminate) + " "
	suffix = " " + render(dialect, updateTmplSuffix, upd)
	return prefix, suffix
}

type Update struct {
	Table     string
	Where     []Where
	Returning []string
}

func UpdateFromIR(ir_upd *ir.Update, dialect Dialect) *Update {
	var returning []string
	if dialect.Features().Returning {
		returning = ir_upd.Model.SelectRefs()
	}

	if len(ir_upd.Joins) == 0 {
		return &Update{
			Table:     ir_upd.Model.Table,
			Where:     WheresFromIR(ir_upd.Where),
			Returning: returning,
		}
	}

	pk_column := ir_upd.Model.PrimaryKey[0].Column

	sel := render(dialect, selectTmpl, Select{
		From:   ir_upd.Model.Table,
		Fields: []string{pk_column},
		Joins:  JoinsFromIR(ir_upd.Joins),
		Where:  WheresFromIR(ir_upd.Where),
	}, noTerminate)

	return &Update{
		Table:     ir_upd.Model.Table,
		Returning: returning,
		Where: []Where{{
			Left:  pk_column,
			Op:    "IN",
			Right: "(" + sel + ")",
		}},
	}
}
