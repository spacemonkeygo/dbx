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
	deleteTmpl = `
	DELETE FROM {{ .From }}
	{{- if .Where }} WHERE
	{{- range $i, $w := .Where }}{{ if $i }} AND{{ end }} {{ $w.Left }} {{ $w.Op }} {{ $w.Right }}{{ end }}
	{{- end -}}`
)

func RenderDelete(dialect Dialect, ir_del *ir.Delete) string {
	return render(dialect, deleteTmpl, DeleteFromIR(ir_del, dialect))
}

type Delete struct {
	From  string
	Where []Where
}

func DeleteFromIR(ir_del *ir.Delete, dialect Dialect) *Delete {
	if len(ir_del.Joins) == 0 {
		return &Delete{
			From:  ir_del.Model.Table,
			Where: WheresFromIR(ir_del.Where),
		}
	}

	pk := ir_del.Model.PrimaryKey[0].ColumnRef()

	sel := render(dialect, selectTmpl, Select{
		From:   ir_del.Model.Table,
		Fields: []string{pk},
		Joins:  JoinsFromIR(ir_del.Joins),
		Where:  WheresFromIR(ir_del.Where),
	}, noTerminate)

	return &Delete{
		From: ir_del.Model.Table,
		Where: []Where{{
			Left:  pk,
			Op:    "IN",
			Right: "(" + sel + ")",
		}},
	}
}
