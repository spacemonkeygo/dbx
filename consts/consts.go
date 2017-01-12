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

package consts

type JoinType int

const (
	InnerJoin JoinType = iota
)

type Operator string

const (
	LT   Operator = "<"
	LE   Operator = "<="
	GT   Operator = ">"
	GE   Operator = ">="
	EQ   Operator = "="
	NE   Operator = "!="
	Like Operator = "like"
)

type FieldType int

const (
	SerialField FieldType = iota
	Serial64Field
	IntField
	Int64Field
	UintField
	Uint64Field
	FloatField
	Float64Field
	TextField
	BoolField
	TimestampField
	TimestampUTCField
	BlobField
)

func (f FieldType) String() string {
	switch f {
	case SerialField:
		return "serial"
	case Serial64Field:
		return "serial64"
	case IntField:
		return "int"
	case Int64Field:
		return "int64"
	case UintField:
		return "uint"
	case Uint64Field:
		return "uint64"
	case FloatField:
		return "float"
	case Float64Field:
		return "float64"
	case TextField:
		return "text"
	case BoolField:
		return "bool"
	case TimestampField:
		return "timestamp"
	case TimestampUTCField:
		return "utimestamp"
	case BlobField:
		return "blob"
	default:
		return "<UNKNOWN-FIELD>"
	}
}

func (f FieldType) AsLink() FieldType {
	switch f {
	case SerialField:
		return IntField
	case Serial64Field:
		return Int64Field
	default:
		return f
	}
}

type RelationKind int

const (
	SetNull RelationKind = iota
	Cascade
	Restrict
)
