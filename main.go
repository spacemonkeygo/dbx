package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/azdagron/dbx/internal/dbx"
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

	var loader dbx.Loader
	if tmpldir != "" {
		loader = DirLoader(tmpldir)
	} else {
		loader = dbx.LoaderFunc(templates.Asset)
	}

	sql, err := dbx.NewSQL(loader, "postgres.tmpl")
	if err != nil {
		return err
	}

	lang, err := dbx.NewLanguage(loader, "golang.tmpl", sql)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	buf.Reset()
	if err := dbx.RenderSchema(schema, sql, &buf); err != nil {
		return err
	}
	os.Stdout.Write(buf.Bytes())

	buf.Reset()
	if err := dbx.RenderCode(schema, lang, &buf); err != nil {
		return err
	}
	// formatted, err := format.Source(buf.Bytes())
	// if err != nil {
	// 	return err
	// }
	// os.Stdout.Write(formatted)

	return nil
}

func loadSchema() *dbx.Schema {
	user := &dbx.Table{Name: "user"}
	user.Columns = []*dbx.Column{
		{
			Table:   user,
			Name:    "pk",
			Type:    "serial64",
			NotNull: true,
		},
		{
			Table:   user,
			Name:    "full_name",
			Type:    "text",
			NotNull: true,
		},
	}
	user.PrimaryKey = user.GetColumns("pk")

	project := &dbx.Table{Name: "project"}
	project.Columns = []*dbx.Column{
		{
			Table:   project,
			Name:    "pk",
			Type:    "serial64",
			NotNull: true,
		},
		{
			Table:   project,
			Name:    "id",
			Type:    "text",
			NotNull: true,
		},
	}
	project.PrimaryKey = project.GetColumns("pk")
	project.Unique = [][]*dbx.Column{
		project.GetColumns("id"),
	}
	project.Queries = []*dbx.Query{
		{},
		{Start: project.GetColumns("id")},
	}

	project_user := &dbx.Table{Name: "project_user"}
	project_user.Columns = []*dbx.Column{
		{
			Table:    project_user,
			Name:     "user_pk",
			Type:     "serial64",
			Relation: user.GetColumn("pk"),
			NotNull:  true,
		},
		{
			Table:    project_user,
			Name:     "project_pk",
			Type:     "serial64",
			Relation: project.GetColumn("pk"),
			NotNull:  true,
		},
	}
	project_user.PrimaryKey = project_user.GetColumns("user_pk", "project_pk")

	user.Queries = append(user.Queries, &dbx.Query{
		Joins: []*dbx.Relation{
			project_user.GetColumn("user_pk").RelationRight(),
			project_user.GetColumn("project_pk").RelationLeft(),
		},
		End: project.GetColumns("pk"),
	})

	project.Queries = append(project.Queries, &dbx.Query{
		Joins: []*dbx.Relation{
			project_user.GetColumn("project_pk").RelationRight(),
			project_user.GetColumn("user_pk").RelationLeft(),
		},
		End: user.GetColumns("pk"),
	})

	bookie := &dbx.Table{Name: "bookie"}
	bookie.Columns = []*dbx.Column{
		{
			Table:   bookie,
			Name:    "pk",
			Type:    "serial64",
			NotNull: true,
		},
		{
			Table:   bookie,
			Type:    "text",
			Name:    "id",
			NotNull: true,
		},
		{
			Table:    bookie,
			Type:     "text",
			Name:     "project_id",
			Relation: project.GetColumn("id"),
			NotNull:  true,
		},
	}
	bookie.PrimaryKey = bookie.GetColumns("pk")
	bookie.Unique = [][]*dbx.Column{
		bookie.GetColumns("id"),
	}
	bookie.Queries = []*dbx.Query{
		{Start: bookie.GetColumns("pk")},
		{Start: bookie.GetColumns("id")},
		{
			Start: bookie.GetColumns("pk"),
			Joins: []*dbx.Relation{
				bookie.GetColumn("project_id").RelationLeft(),
			},
			End: project.GetColumns("pk"),
		},
	}

	billing_key := &dbx.Table{Name: "billing_key"}
	billing_key.Columns = []*dbx.Column{
		{
			Table:   billing_key,
			Name:    "pk",
			Type:    "serial64",
			NotNull: true,
		},
		{
			Table:   billing_key,
			Name:    "id",
			Type:    "text",
			NotNull: true,
		},
		{
			Table:    billing_key,
			Name:     "bookie_id",
			Type:     "text",
			Relation: bookie.GetColumn("id"),
			NotNull:  true,
		},
	}
	billing_key.PrimaryKey = billing_key.GetColumns("pk")
	billing_key.Unique = [][]*dbx.Column{
		billing_key.GetColumns("id"),
	}
	billing_key.Queries = []*dbx.Query{
		{Start: billing_key.GetColumns("pk")},
		{Start: billing_key.GetColumns("id")},
		{
			Joins: []*dbx.Relation{
				billing_key.GetColumn("bookie_id").RelationLeft(),
				bookie.GetColumn("project_id").RelationLeft(),
			},
			End: project.GetColumns("pk"),
		},
		{
			Joins: []*dbx.Relation{
				billing_key.GetColumn("bookie_id").RelationLeft(),
			},
			End: bookie.GetColumns("id"),
		},
	}

	schema := &dbx.Schema{
		Tables: []*dbx.Table{
			user,
			project,
			project_user,
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
