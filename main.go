package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/spacemonkeygo/dbx/internal/dbx"
	"github.com/spacemonkeygo/dbx/internal/dbx/dbxdialect"
	"github.com/spacemonkeygo/dbx/internal/dbx/dbxlanguage"
	"github.com/spacemonkeygo/dbx/templates"
)

func main() {
	app := cli.App("dbx", "generate SQL schema and matching code")

	template_dir_arg := app.StringOpt("t templates", "", "templates directory")
	in_arg := app.StringArg("IN", "", "path to the yaml description")
	out_arg := app.StringArg("OUT", "", "output file (- for stdout)")

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
			dialect, err := dbxdialect.NewPostgres(loader)
			die(err)
			die(generateSQLSchema(*out_arg, schema, dialect))
		}
	})

	app.Command("code", "generate code", func(cmd *cli.Cmd) {
		pkg_name := cmd.StringOpt("p package", "db",
			"package name for generated code")
		format_code := cmd.BoolOpt("f format", true,
			"format the code")
		cmd.Action = func() {
			dialect, err := dbxdialect.NewPostgres(loader)
			die(err)
			lang, err := dbxlanguage.NewGolang(loader, dialect,
				&dbxlanguage.GolangOptions{
					Package: *pkg_name,
				})
			die(err)

			die(generateCode(*out_arg, schema, dialect, lang, *format_code))
		}
	})

	app.Run(os.Args)
}

func generateSQLSchema(out string, schema *dbx.Schema, dialect dbx.Dialect) (
	err error) {

	sql, err := dialect.RenderSchema(schema)
	if err != nil {
		return err
	}
	return writeOut(out, []byte(sql))
}

func generateCode(out string, schema *dbx.Schema, dialect dbx.Dialect,
	lang dbx.Language, format_code bool) (err error) {

	var buf bytes.Buffer
	if err := dbx.RenderCode(&buf, schema, dialect, lang); err != nil {
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
