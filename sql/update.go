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

package sql

import "gopkg.in/spacemonkeygo/dbx.v1/ir"

const (
	updateTmplPrefix = `
	UPDATE {{ .Table }} SET`
	updateTmplSuffix = `
	{{ if .Where }} WHERE {{- range $i, $w := .Where }}{{ if $i }} AND{{ end }} {{ $w.Left }} {{ $w.Op }} {{ $w.Right }}{{ end }} {{- end -}}
	{{ if .Returning }} RETURNING *{{ end }}`
)

func RenderUpdate(dialect Dialect, ir_upd *ir.Update) (prefix, suffix string) {
	upd := UpdateFromIR(ir_upd, dialect)
	return render(updateTmplPrefix, upd, noTerminate) + " ",
		" " + render(updateTmplSuffix, upd)
}

type Update struct {
	Table     string
	Where     []Where
	Returning bool
}

func UpdateFromIR(ir_upd *ir.Update, dialect Dialect) *Update {
	return &Update{
		Table:     ir_upd.Model.TableName(),
		Where:     WheresFromIR(ir_upd.Where),
		Returning: dialect.Features().Returning,
	}
}
