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

package ast

import (
	"fmt"
	"text/scanner"
)

type FieldRef struct {
	Pos   scanner.Position
	Model string
	Field string
}

func (r *FieldRef) String() string {
	if r.Field == "" {
		return r.Model
	}
	if r.Model == "" {
		return r.Field
	}
	return fmt.Sprintf("%s.%s", r.Model, r.Field)
}

func (f *FieldRef) Relative() *RelativeFieldRef {
	return &RelativeFieldRef{
		Pos:   f.Pos,
		Field: f.Field,
	}
}

type RelativeFieldRef struct {
	Pos   scanner.Position
	Field string
}

func (r *RelativeFieldRef) String() string { return r.Field }
