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

//go:generate go run gen_bundle.go

package sqlbundle

const Source = "type sqlgen_SQL interface {\n\tRender() string\n\n\tprivate()\n}\n\ntype sqlgen_Dialect interface {\n\tRebind(sql string) string\n}\n\ntype sqlgen_RenderOp int\n\nconst (\n\tsqlgen_NoFlatten sqlgen_RenderOp = iota\n\tsqlgen_NoTerminate\n)\n\nfunc sqlgen_Render(dialect sqlgen_Dialect, sql sqlgen_SQL, ops ...sqlgen_RenderOp) string {\n\tout := sql.Render()\n\n\tflatten := true\n\tterminate := true\n\tfor _, op := range ops {\n\t\tswitch op {\n\t\tcase sqlgen_NoFlatten:\n\t\t\tflatten = false\n\t\tcase sqlgen_NoTerminate:\n\t\t\tterminate = false\n\t\t}\n\t}\n\n\tif flatten {\n\t\tout = sqlgen_flattenSQL(out)\n\t}\n\tif terminate {\n\t\tout += \";\"\n\t}\n\n\treturn dialect.Rebind(out)\n}\n\nvar sqlgen_reSpace = regexp.MustCompile(`\\s+`)\n\nfunc sqlgen_flattenSQL(s string) string {\n\treturn strings.TrimSpace(sqlgen_reSpace.ReplaceAllString(s, \" \"))\n}\n\ntype sqlgen_postgres struct{}\n\nfunc sqlgen_Postgres() sqlgen_Dialect {\n\treturn &sqlgen_postgres{}\n}\n\nfunc (p *sqlgen_postgres) Rebind(sql string) string {\n\tout := make([]byte, 0, len(sql)+10)\n\n\tj := 1\n\tfor i := 0; i < len(sql); i++ {\n\t\tch := sql[i]\n\t\tif ch != '?' {\n\t\t\tout = append(out, ch)\n\t\t\tcontinue\n\t\t}\n\n\t\tout = append(out, '$')\n\t\tout = append(out, strconv.Itoa(j)...)\n\t\tj++\n\t}\n\n\treturn string(out)\n}\n\ntype sqlgen_sqlite3 struct{}\n\nfunc sqlgen_SQLite3() sqlgen_Dialect {\n\treturn &sqlgen_sqlite3{}\n}\n\nfunc (s *sqlgen_sqlite3) Rebind(sql string) string {\n\treturn sql\n}\n\ntype sqlgen_Literal string\n\nfunc (sqlgen_Literal) private() {}\n\nfunc (l sqlgen_Literal) Render() string { return string(l) }\n\ntype sqlgen_Literals struct {\n\tJoin string\n\tSQLs []sqlgen_SQL\n}\n\nfunc (sqlgen_Literals) private() {}\n\nfunc (l sqlgen_Literals) Render() string {\n\tvar out bytes.Buffer\n\n\tfirst := true\n\tfor _, sql := range l.SQLs {\n\t\tif sql == nil {\n\t\t\tcontinue\n\t\t}\n\t\tif !first {\n\t\t\tout.WriteString(l.Join)\n\t\t}\n\t\tfirst = false\n\t\tout.WriteString(sql.Render())\n\t}\n\n\treturn out.String()\n}\n\ntype sqlgen_Hole struct {\n\tName string\n\n\tval sqlgen_SQL\n}\n\nfunc (*sqlgen_Hole) private() {}\n\nfunc (h *sqlgen_Hole) Fill(sql sqlgen_SQL) { h.val = sql }\n\nfunc (h *sqlgen_Hole) Render() string {\n\tif h.val == nil {\n\t\treturn \"\"\n\t}\n\treturn h.val.Render()\n}"
