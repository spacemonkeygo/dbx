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

// DBX implements code generation for database schemas and accessors.
package main // import "gopkg.in/spacemonkeygo/dbx.v1/bin/dbx"

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/code"
	"gopkg.in/spacemonkeygo/dbx.v1/code/golang"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/parser"
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
	pubtemplates "gopkg.in/spacemonkeygo/dbx.v1/templates"
	"gopkg.in/spacemonkeygo/dbx.v1/tmplutil"
)

func main() {
	app := cli.App("dbx", "generate SQL schema and matching code")

	template_dir_arg := app.StringOpt("t templates", "", "templates directory")
	in_arg := app.StringArg("IN", "", "path to the description")
	out_arg := app.StringArg("OUT", "", "output file (- for stdout)")
	dialects_opt := app.StringsOpt("d dialect", nil,
		"SQL dialect to use (postgres if unspecified)")

	var err error
	die := func(err error) {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			cli.Exit(1)
		}
	}

	var ast_root *ast.Root
	var root *ir.Root
	var loader tmplutil.Loader

	app.Before = func() {
		ast_root, err = parser.ParseFile(*in_arg)
		die(err)
		root, err = ir.Transform(ast_root)
		die(err)
		err = ir.GenerateBasicQueries(root, ir.GenerateOptions{
			Insert:             true,
			RawInsert:          true,
			SelectAll:          true,
			SelectByPrimaryKey: true,
			SelectByUnique:     true,
			DeleteByPrimaryKey: true,
			DeleteByUnique:     true,
			UpdateByPrimaryKey: true,
			UpdateByUnique:     true,
			Count:              true,
		})
		die(err)

		if *template_dir_arg != "" {
			loader = tmplutil.DirLoader(*template_dir_arg)
		} else {
			loader = tmplutil.BinLoader(pubtemplates.Asset)
		}
	}

	app.Command("schema", "generate SQL schema", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			//dialects, err := createDialects(*dialects_opt)
			//die(err)
			//die(generateSQLSchema(*out_arg, dialects, root))
		}
	})

	app.Command("code", "generate code", func(cmd *cli.Cmd) {
		pkg_name := cmd.StringOpt("p package", "db",
			"package name for generated code")
		cmd.Action = func() {
			dialects, err := createDialects(*dialects_opt)
			die(err)
			renderer, err := golang.New(loader, &golang.Options{
				Package: *pkg_name,
			})
			die(err)

			die(generateCode(*out_arg, root, dialects, renderer))
		}
	})

	app.Run(os.Args)
}

func createDialects(which []string) (out []sql.Dialect, err error) {
	if len(which) == 0 {
		which = append(which, "postgres")
	}
	for _, name := range which {
		var d sql.Dialect
		switch name {
		case "postgres":
			d = sql.Postgres()
		case "sqlite3":
			d = sql.SQLite3()
		default:
			return nil, fmt.Errorf("unknown dialect %q", name)
		}
		out = append(out, d)
	}
	return out, nil
}

//func generateSQLSchema(out string, dialects []sql.Dialect, root *ir.Root) (
//	err error) {
//
//	for _, dialect := range dialects {
//		sql, err := dialect.RenderSchema(schema)
//		if err != nil {
//			return err
//		}
//		if err = writeOut(out+"-"+dialect.Name(), []byte(sql)); err != nil {
//			return err
//		}
//	}
//	return nil
//}

func generateCode(out string, root *ir.Root, dialects []sql.Dialect,
	renderer code.Renderer) (err error) {

	rendered, err := renderer.RenderCode(root, dialects)
	if err != nil {
		return err
	}

	return writeOut(out, rendered)
}

func writeOut(out string, data []byte) (err error) {
	w := os.Stdout
	if out != "-" {
		w, err = os.Create(out)
		if err != nil {
			return fmt.Errorf("unable to open output file: %s", err)
		}
		defer w.Close()
	}
	_, err = w.Write(data)
	return err
}
