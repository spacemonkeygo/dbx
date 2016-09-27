package dbx

import "io"

func RenderSchema(schema *Schema, sql *SQL, w io.Writer) (err error) {
	data, err := sql.RenderSchema(schema)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, data)
	return err
}
