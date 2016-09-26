package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/azdagron/dbx/internal"
	"github.com/azdagron/dbx/templates"
	cli "github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("dbx", "generate SQL schema and matching code")

	tmpldir := app.StringOpt("t templates", "", "templates directory")

	app.Action = func() {
		if err := run(*tmpldir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			cli.Exit(1)
		}
	}

	app.Run(os.Args)
}

func run(tmpldir string) (err error) {
	schema := loadSchema()

	var loader internal.Loader
	if tmpldir != "" {
		loader = DirLoader(tmpldir)
	} else {
		loader = internal.LoaderFunc(templates.Asset)
	}

	sql, err := internal.NewSQL(loader, "postgres.tmpl")
	if err != nil {
		return err
	}

	lang, err := internal.NewLanguage(loader, "golang.tmpl", sql)
	if err != nil {
		return err
	}

	if err := internal.RenderSchema(schema, sql, os.Stdout); err != nil {
		return err
	}
	if err := internal.RenderCode(schema, lang, os.Stdout); err != nil {
		return err
	}
	return nil
}

func loadSchema() *internal.Schema {
	project := &internal.Table{Name: "project"}
	project.Columns = append(project.Columns, &internal.Column{
		Table:      project,
		Name:       "pk",
		Type:       "serial64",
		PrimaryKey: true,
	}, &internal.Column{
		Table:  project,
		Name:   "id",
		Type:   "text",
		Unique: true,
	})

	bookie := &internal.Table{Name: "bookie"}
	bookie.Columns = append(bookie.Columns, &internal.Column{
		Table:      bookie,
		Name:       "pk",
		Type:       "serial64",
		PrimaryKey: true,
	}, &internal.Column{
		Table:  bookie,
		Type:   "text",
		Name:   "id",
		Unique: true,
	}, &internal.Column{
		Table:    bookie,
		Type:     "text",
		Name:     "project_id",
		Relation: project.Column("id"),
	})

	billing_key := &internal.Table{Name: "billing_key"}
	billing_key.Columns = append(billing_key.Columns, &internal.Column{
		Table:      billing_key,
		Name:       "pk",
		Type:       "serial64",
		PrimaryKey: true,
	}, &internal.Column{
		Table:  billing_key,
		Name:   "id",
		Type:   "text",
		Unique: true,
	}, &internal.Column{
		Table:    billing_key,
		Name:     "bookie_id",
		Type:     "text",
		Relation: bookie.Column("id"),
	})

	schema := &internal.Schema{
		Tables: []*internal.Table{
			project,
			bookie,
			billing_key,
		},
	}

	return schema
}

type DirLoader string

func (d DirLoader) Load(name string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(string(d), name))
}
