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

package code

import (
	"io"

	"github.com/spacemonkeygo/errors"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
)

var (
	Error = errors.NewClass("code")
)

type Language interface {
	Format([]byte) ([]byte, error)
	RenderHeader(w io.Writer, root *ir.Root, dialects []sql.Dialect) error
	RenderInsert(w io.Writer, model *ir.Model, dialect sql.Dialect) error
	RenderSelect(w io.Writer, sel *ir.Select, dialect sql.Dialect) error
	RenderDelete(w io.Writer, del *ir.Delete, dialect sql.Dialect) error
	RenderFooter(w io.Writer, root *ir.Root, dialects []sql.Dialect) error
}
