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

//
// DO NOT EDIT: automatically generated code.
//

//go:generate go run gen_bundle.go

package sqlbundle

const (
	Source = "type __sqlbundle_SQL interface {\n\tRender() string\n\n\tprivate()\n}\n\ntype __sqlbundle_Dialect interface {\n\tRebind(sql string) string\n}\n\ntype __sqlbundle_RenderOp int\n\nconst (\n\t__sqlbundle_NoFlatten __sqlbundle_RenderOp = iota\n\t__sqlbundle_NoTerminate\n)\n\nfunc __sqlbundle_Render(dialect __sqlbundle_Dialect, sql __sqlbundle_SQL, ops ...__sqlbundle_RenderOp) string {\n\tout := sql.Render()\n\n\tflatten := true\n\tterminate := true\n\tfor _, op := range ops {\n\t\tswitch op {\n\t\tcase __sqlbundle_NoFlatten:\n\t\t\tflatten = false\n\t\tcase __sqlbundle_NoTerminate:\n\t\t\tterminate = false\n\t\t}\n\t}\n\n\tif flatten {\n\t\tout = __sqlbundle_flattenSQL(out)\n\t}\n\tif terminate {\n\t\tout += \";\"\n\t}\n\n\treturn dialect.Rebind(out)\n}\n\nvar __sqlbundle_reSpace = regexp.MustCompile(`\\s+`)\n\nfunc __sqlbundle_flattenSQL(s string) string {\n\treturn strings.TrimSpace(__sqlbundle_reSpace.ReplaceAllString(s, \" \"))\n}\n\ntype __sqlbundle_postgres struct{}\n\nfunc __sqlbundle_Postgres() __sqlbundle_Dialect {\n\treturn &__sqlbundle_postgres{}\n}\n\nfunc (p *__sqlbundle_postgres) Rebind(sql string) string {\n\tout := make([]byte, 0, len(sql)+10)\n\n\tj := 1\n\tfor i := 0; i < len(sql); i++ {\n\t\tch := sql[i]\n\t\tif ch != '?' {\n\t\t\tout = append(out, ch)\n\t\t\tcontinue\n\t\t}\n\n\t\tout = append(out, '$')\n\t\tout = append(out, strconv.Itoa(j)...)\n\t\tj++\n\t}\n\n\treturn string(out)\n}\n\ntype __sqlbundle_sqlite3 struct{}\n\nfunc __sqlbundle_SQLite3() __sqlbundle_Dialect {\n\treturn &__sqlbundle_sqlite3{}\n}\n\nfunc (s *__sqlbundle_sqlite3) Rebind(sql string) string {\n\treturn sql\n}\n\ntype __sqlbundle_Literal string\n\nfunc (__sqlbundle_Literal) private() {}\n\nfunc (l __sqlbundle_Literal) Render() string { return string(l) }\n\ntype __sqlbundle_Literals struct {\n\tJoin string\n\tSQLs []__sqlbundle_SQL\n}\n\nfunc (__sqlbundle_Literals) private() {}\n\nfunc (l __sqlbundle_Literals) Render() string {\n\tvar out bytes.Buffer\n\n\tfirst := true\n\tfor _, sql := range l.SQLs {\n\t\tif sql == nil {\n\t\t\tcontinue\n\t\t}\n\t\tif !first {\n\t\t\tout.WriteString(l.Join)\n\t\t}\n\t\tfirst = false\n\t\tout.WriteString(sql.Render())\n\t}\n\n\treturn out.String()\n}\n\ntype __sqlbundle_Condition struct {\n\tName string\n\n\tField string\n\tEqual bool\n\tNull  bool\n}\n\nfunc (*__sqlbundle_Condition) private() {}\n\nfunc (c *__sqlbundle_Condition) Render() string {\n\n\tfield := c.Field\n\tif field != \"\" {\n\t\tfield += \" \"\n\t}\n\n\tswitch {\n\tcase c.Equal && c.Null:\n\t\treturn field + \"is null\"\n\tcase c.Equal && !c.Null:\n\t\treturn field + \"= ?\"\n\tcase !c.Equal && c.Null:\n\t\treturn field + \"is not null\"\n\tcase !c.Equal && !c.Null:\n\t\treturn field + \"!= ?\"\n\tdefault:\n\t\tpanic(\"unhandled case\")\n\t}\n}"
	Prefix = "__sqlbundle_"
)
