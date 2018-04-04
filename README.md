# DBX

DBX is a tool to generate database schemas and code to operate with it. It
currently generates Go bindings to Postgres and/or SQLite, but it should be
fairly straightforward to add other database *and* language targets.

## How it works

DBX takes a description of models and operations to perform on those models
and can generate code to interact with sql databases.

## Installing

```
go get gopkg.in/spacemonkeygo/dbx.v1
```

## Basic Example

### Declaring a Model

Consider a basic `user` model with a primary key, a unique identifier, some
timestamps for keeping track of modifications and a name. We will require that
the id and the name fields are unique.

```
model user (
	key    pk
	unique id
	unique name

	field pk         serial64
	field created_at timestamp ( autoinsert )
	field updated_at timestamp ( autoinsert, autoupdate )
	field id         text
	field name       text
)
```

If we place this model in a file called `example.dbx`, we can build some go
source with the command

```
$ dbx.v1 golang example.dbx .
```

This will create an `example.go` file in the current directory. Check the
output of `dbx.v1 golang` for more options like controling the package name or
other features of the generated code.

Generating a schema is also straightforward:

```
$ dbx.v1 schema examples.dbx .
```

This creates an `example.dbx.postgres.sql` file in the current directory with
sql statements to create the tables for the models.

By default DBX will generate code for all of the models and fields and use the
postgres SQL dialect. See the dialects section below for more discussion on
other supported dialects and how to generate them.

This example package doesn't do very much because we didn't ask for very much,
but it does include a struct definition like

```
type User struct {
	Pk        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	Id        string
	Name      string
}
```

as well as concrete types `DB` and `Tx`, and interfaces that they implement
that look like

```
type Methods interface {
}

type TxMethods interface {
	Methods

	Commit() error
	Rollback() error
}

type DBMethods interface {
	Schema() string
	Methods
}
```

The `Methods` interface is shared between the `Tx` and `DB` interfaces and will
contain methods to interact with the database when they are generated. If you
were to pass the userdata option on the generate command, then the `User`
struct would come with an `interface{}` and a `sync.Mutex` to store some
arbitrary data on a value.

The package comes with some customizable hooks.

```
var WrapErr = func(err *Error) error { return err }
var Logger func(format string, args ...interface{})
```

- All of the errors returned by the database are passed through the `WrapErr`
function so that you may process them however you wish: by adding contextual
information or stack traces for example.
- If the `Logger` is not nil, all of the SQL statements that would be executed
are passed to it in the args, as well as other informational statements.
- There is a `Hooks` type on the `*DB` that contains hooks like `Now` for
mocking out time in your tests so that any `autoinsert`/`autoupdate` time
fields can be given a deterministic value.

The package has an `Open` function that returns a `*DB` instance. It's
signature looks like

```
func Open(driver, source string) (db *DB, err error)
```

The driver must be one of the dialects passed in at generation time, which by
default is just `postgres`. The `*DB` type lets you `Open` a new transaction
represented by `*Tx`, `Close` the database, or run queries as normal. It has a
`DB` field that exposes the raw `"database/sql".(*DB)` value.

We can instruct DBX to generate code for interacting with the database now.

### Declaring Operations

There are four kinds of operations, `create`, `read`, `update` and `delete`. We
can add one of each operation for the `user` model based on the primary key:

```
create user ( )
update user ( where user.pk = ? )
delete user ( where user.pk = ? )
read one (
	select user
	where  user.pk = ?
)
```

Regenerating the Go code will expand our database interface:

```
type Methods interface {
	Create_User(ctx context.Context,
		user_id User_Id_Field,
		user_name User_Name_Field) (
		user *User, err error)

	Delete_User_By_Pk(ctx context.Context,
		user_pk User_Pk_Field) (
		deleted bool, err error)

	Get_User_By_Pk(ctx context.Context,
		user_pk User_Pk_Field) (
		user *User, err error)

	Update_User_By_Pk(ctx context.Context,
		user_pk User_Pk_Field,
		update User_Update_Fields) (
		user *User, err error)
}
```

The fields are all wrapped in their own type so that arguments cannot be passed
in the wrong order: both the id and name fields are strings, and so we prevent
any of those errors at compile time.

For example, to create a user, we could write

```
db.Create_User(ctx,
		User_Id("some unique id i just generated"),
		User_Name("Donny B. Xavier"))
```

### Transactions

DBX attempts to expose transaction handling, just like the `database/sql`
package, but that can sometimes be verbose with handling Commits and Rollbacks.
Consider a function to create a user within a transaction:

```
func createUser(ctx context.Context, db *DB) (user *User, err error) {
	tx, err := db.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			// tx.Rollback() returns an error, perhaps we should log it, or
			// do something else? the choice is yours.
			tx.Rollback()
		}
	}()

	return tx.Create_User(ctx, 
		User_Id("some unique id i just generated"),
		User_Name("Donny B. Xavier"))
}
```

Go allows you to define a package as a collection of multiple files, and so it
might be worthwhile for you to add a helper method to the `*DB` type in another
file like this:

```
func (db *DB) WithTx(ctx context.Context,
	fn func(context.Context, *Tx) error) (err error) {

	tx, err := db.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback() // log this perhaps?
		}
	}()
	return fn(ctx, tx)
}
```

Then `createUser` can be succinctly written

```
func createUser(ctx context.Context, db *DB) (user *User, err error) {
	err = db.WithTx(func(ctx context.Context, tx *Tx) error) {
		user, err = tx.Create_User(ctx,
			User_Id("some unique id i just generated"),
			User_Name("Donny B. Xavier"))
		return err
	})
	return user, err
}
```

DBX does not generate this helper for you so that you can have full control
over how you want to handle the error in the Rollback case.

### Dialects

DBX doesn't work with just Postgres, and is designed to be agnostic to many
different database engines. Currently, it supports Postgres and SQLite3. Any
of the above commands can be passed the `--dialect` (or, shorthand `-d`) flag
to specify additional dialects. For example, running

```
dbx.v1 schema -d postgres -d sqlite3 example.dbx .
dbx.v1 golang -d postgres -d sqlite3 example.dbx .
```

will create both `example.dbx.postgres.sql` and `example.dbx.sqlite3.sql` with
the statements required to create the tables, and generate the Go code to
operate with both sqlite3 and postgres.

### Generate

All of these commands are intended to normally be used with `//go:generate`
directives, such as:

```
//go:generate dbx.v1 golang -d postgres -d sqlite3 example.dbx .
//go:generate dbx.v1 schema -d postgres -d sqlite3 example.dbx .
```

A great spot to put them would be in the file that modifies the hooks and adds
other customizations.

## Details

Detailed documentation below. If you notice any difference between the
documentation and the actual behavior, please open an issue and we'll fix it!

### Grammar

A DBX file has two constructs: tuples and lists. A list contains comma
separated tuples, and tuples contain white space separated strings or more
lists. Somewhat like Go automatic semicolon insertion, commas are inserted at a
newline if the previous token was not a comma.

For example, this is a list of three tuples:

```
(
	tuple one
	another tuple here
	the third ( tuple )
)
```

The first tuple contains two strings, `"tuple"` and `"one"`. The second tuple
contains three strings, `"another"`, `"tuple"`, and `"here"`. The last tuple
contains two strings and a list containing one tuple, `"the"`, `"third"` and
`( tuple )`. This list could be written with explicit commas either with or
without newlines:

```
( tuple one, another tuple here, the third (
	tuple
) )
```

```
( tuple one,
  another tuple here,
  the third ( tuple ),
)
```

are all the same grammatically. A dbx file implicitly has a list at the top
level that does not require opening and closing parenthesis.

### Models

```
model <name> (
	// table is optional and gives the name of the table the model will use.
	table <name>

	// key is required and declares the primary key for the model. it can
	// be either a single field or a multiple fields for a composite primary
	// key.
	key <field names>
	
	// unique constraints are optional and on any number of fields. you can
	// have as many unique constraints as you want.
	unique <field names>
	
	// indexes are optional and you can have as many as you want.
	index (
		// the name of the index.
		// BUG: we only allow one empty name index :)
		name <name>
		
		// fields describes which fields are in the index
		fields <fields>
		
		// when set, the index will have a unique constraint
		unique
	)
	
	// field declares a normal field to have the name and type. attributes is
	// an optional list that can be used to tune specific details about the
	// field like nullable. see the section on attributes to see the full list.
	field <name> <type> ( attributes )
	
	// a model can have foreign key relations to another model's field. the
	// relation describes what happens on delete: if the related field's row is
	// removed, what do we do to the row that describes this model? as normal
	// fields, there are a number of optional attributes.
	field <name> <model>.<field> <relation kind> ( attributes )
)
```

#### Attributes

Fields can have these attributes

- `column <name>`: use this name for the column name
- `nullable`: this field is nullable (can have NULL as a value)
- `updatable`: this field can be updated
- `autoinsert`: this field will be inserted with the zero value of the type, or
the current time if the field is a time field, and you won't have to specify it
on any Create calls.
- `autoupdate`: this field will be updated with the zero value of the type, or
the current time if the field is a time field, and you won't have to specify it
on any Update calls. BUG: this is only really useful on timestamp fields :)
- `length <length>`: on text fields, this specifies the maximum length of the
text.

#### Field Types

Fields may be any of these types

- `serial`
- `serial64`
- `int`
- `int64`
- `uint`
- `uint64`
- `bool`
- `text`
- `timestamp`
- `utimestamp` (timestamp with no timezone. expected to be in UTC)
- `float`
- `float64`
- `blob`

#### Foreign Key Relation Kinds

A foreign key relation can be any of these

- `setnull`: when the related row goes away, set this field to null. the field
must be `nullable`.
- `cascade`: when the related row goes away, delete this row.
- `restrict`: do not allow the related row to go away.

#### Foreign Key Attributes

A foreign key can have these attributes

- `column <name>`: use this name for the column name
- `nullable`: this field is nullable (can have NULL as a value)
- `updatable`: this field can be updated

### Create

```
create <model> (
	// raw will cause the generation of a "raw" create that exposes every field
	raw
	
	// suffix will cause the generated create method to have the desired value
	suffix <parts>
)
```

### Read


`<views>` is a list of views that describe what kind of reads to generate and
is constrained by whether or not a read is distinct. a read is said to be
distinct if the where clauses and join conditions identify a unique result.

the following views are defined for all reads:
* `count`  - returns the number of results
* `has`    - returns if there are results or not
* `first`  - returns the first result or nothing
* `scalar` - returns a single result, nothing, or fails if there is more than
		     one result
* `one`    - returns a single result or fails if there are no results or more
             than one result

the following views are only defined for non-distinct reads:
* `all` - returns all results
* `limitoffset` - returns a limited number of results starting at an offset
* `paged` - returns limited number of results paged by a forward iterator

```
read <views> (
	// select describes what values will be returned from the read. you can
	// specify either models or just a field of a model, like "user" or
	// "project.id"
	select <field refs>
	
	// a read can have any number of where clauses. the clause refers to an
	// expression, an operation like "!=" or "<=", and another expression. if
	// the right side field has a placeholder (?), the read will fill it in
	// with an argument and be generated with a parameter for that argument. 
	// multiple where clauses will be joined by "and"s.
	//
	// <expr> can be one of the following:
	// 1) placeholder
	//    where animal.name = ?
	// 2) null
	//    where animal.name = null
	// 3) string literal
	//    where animal.name = "Tiger"
	// 4) number literal
	//    where animal.age < 30
	// 5) boolean literal
	//    where animal.dead = false
	// 6) model field reference: <model>.<field>
	//    where animal.height = animal.width
	// 7) SQL function call: <name>(<expr>)
	//    where lower(animal.name) = "tiger"
	//
	// SQL function calls take an expression for each argument. Currently only
	// "lower" is implemented.
	//
	// <limited-expr> is the same as <expr> except that it can only contain
	// a model field reference, optionally wrapped in one or more function
	// calls.
	where <limited-expr> <op> <expr>
	
	// a join describes a join for the read. it brings the right hand side
	// model into scope for the selects, and the joins must be in a consistent
	// order.
	join <model.field> = <model.field>
	
	// orderby controls the order the rows are returned. direction has to be
	// either "asc" or "desc".
	orderby <direction> <model.field>
	
	// suffix will cause the generated read methods to have the desired value
	suffix <parts>
)
```

### Update

See the documentation on Read for information about where join, and suffix.
Update is required to have enough information in the where and join clauses
for dbx to determine that it will be updating a single row.

```
update <model> (
	where <model.field> <op> <model.field or "?">
	join <model.field> = <model.field>
	suffix <parts>
)
```

### Delete

See the documentation on Read for information about where, join and suffix.

```
delete <model> (
	where <model.field> <op> <model.field or "?">
	join <model.field> = <model.field>
	suffix <parts>
)
```

## Other

### Formatting

DBX comes with a formatter for your dbx source code. It currently has some bugs
and limitations, but defines a canonical way to store your dbx files.

- It ignores comments, so formatting will remove any comments :)
- The command can only read from stdin, and output to stdout.

### Errors

DBX takes errors very seriously and attempts to have a great user experience
around them. If you think an error is misleading, the caret is in the wrong or
less than optimal position, or not obvious what the solution is, please open an
issue and we will try to explain and make the error better. For example,
passing the model with a typo for the foreign key:

```
model user (
	key   pf
	field pk serial64
)
```

we receive the error:

```
example.dbx:2:8: no field "pf" defined on model "user"

context:
   1: model user (
   2:     key   pf
                ^
```


## License

Copyright (C) 2017 Space Monkey, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
