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
package main // import "gopkg.in/spacemonkeygo/dbx.v0/bin/dbx"

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"gopkg.in/spacemonkeygo/dbx.v0/internal/dbx"
	"gopkg.in/spacemonkeygo/dbx.v0/internal/dbx/dialect"
	"gopkg.in/spacemonkeygo/dbx.v0/internal/dbx/language"
	"gopkg.in/spacemonkeygo/dbx.v0/templates"
)

func main() {
	app := cli.App("dbx", "generate SQL schema and matching code")

	template_dir_arg := app.StringOpt("t templates", "", "templates directory")
	in_arg := app.StringArg("IN", "", "path to the yaml description")
	out_arg := app.StringArg("OUT", "", "output file (- for stdout)")
	dialects_opt := app.StringsOpt("d dialect", nil,
		"SQL dialect to use")

	var err error
	die := func(err error) {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			cli.Exit(1)
		}
	}

	var schema *dbx.Schema
	var loader dbx.Loader

	app.Before = func() {
		schema, err = dbx.LoadSchema(*in_arg)
		die(err)

		if *template_dir_arg != "" {
			loader = dbx.DirLoader(*template_dir_arg)
		} else {
			loader = dbx.BinLoader(templates.Asset)
		}
	}

	app.Command("schema", "generate SQL schema", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			dialects, err := createDialects(*dialects_opt, loader)
			die(err)
			die(generateSQLSchema(*out_arg, dialects, schema))
		}
	})

	app.Command("code", "generate code", func(cmd *cli.Cmd) {
		pkg_name := cmd.StringOpt("p package", "db",
			"package name for generated code")
		format_code := cmd.BoolOpt("f format", true,
			"format the code")
		cmd.Action = func() {
			dialects, err := createDialects(*dialects_opt, loader)
			die(err)
			lang, err := language.NewGolang(loader,
				&language.GolangOptions{
					Package: *pkg_name,
				})
			die(err)

			die(generateCode(*out_arg, schema, dialects, lang, *format_code))
		}
	})

	app.Run(os.Args)
}

func createDialects(which []string, loader dbx.Loader) (
	out []dbx.Dialect, err error) {

	for _, name := range which {
		var d dbx.Dialect
		var err error
		switch name {
		case "postgres":
			d, err = dialect.NewPostgres(loader)
		case "sqlite3":
			d, err = dialect.NewSQLite3(loader)
		default:
			return nil, fmt.Errorf("unknown dialect %q", name)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func generateSQLSchema(out string, dialects []dbx.Dialect, schema *dbx.Schema) (
	err error) {

	for _, dialect := range dialects {
		sql, err := dialect.RenderSchema(schema)
		if err != nil {
			return err
		}
		if err = writeOut(out+"-"+dialect.Name(), []byte(sql)); err != nil {
			return err
		}
	}
	return nil
}

func generateCode(out string, schema *dbx.Schema, dialects []dbx.Dialect,
	lang dbx.Language, format_code bool) (err error) {

	var buf bytes.Buffer
	if err := dbx.RenderCode(&buf, schema, dialects, lang); err != nil {
		return err
	}
	rendered := buf.Bytes()

	if format_code {
		formatted, err := lang.Format(rendered)
		if err != nil {
			dumpLinedSource(rendered)
			return err
		}
		rendered = formatted
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

func dumpLinedSource(source []byte) {
	// scan once to find out how many lines
	scanner := bufio.NewScanner(bytes.NewReader(source))
	var lines int
	for scanner.Scan() {
		lines++
	}
	align := 1
	for ; lines > 0; lines = lines / 10 {
		align++
	}

	// now dump with aligned line numbers
	format := fmt.Sprintf("%%%dd: %%s\n", align)
	scanner = bufio.NewScanner(bytes.NewReader(source))
	for i := 1; scanner.Scan(); i++ {
		line := scanner.Text()
		fmt.Fprintf(os.Stderr, format, i, line)
	}
}
