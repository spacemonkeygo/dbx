package dbx

import (
	"io"

	"github.com/spacemonkeygo/errors"
)

var Error = errors.NewClass("dbx")

type Language interface {
	Name() string
	RenderHeader(w io.Writer, schema *Schema) error
	RenderSelect(w io.Writer, sql string, params *SelectParams) error
	RenderCount(w io.Writer, sql string, params *SelectParams) error
	RenderDelete(w io.Writer, sql string, params *DeleteParams) error
	RenderInsert(w io.Writer, sql string, params *InsertParams) error
	Format([]byte) ([]byte, error)
}

type Dialect interface {
	Name() string
	RenderSchema(schema *Schema) (string, error)
	RenderSelect(params *SelectParams) (string, error)
	RenderCount(params *SelectParams) (string, error)
	RenderDelete(params *DeleteParams) (string, error)
	RenderInsert(params *InsertParams) (string, error)
	InsertReturns() bool
}
