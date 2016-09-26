package prettyprint

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"unicode"
)

const (
	indention = "    "
	padding   = " "
)

func New(w io.Writer) *Printer {
	return &Printer{
		Writer:      w,
		ByteEncoder: base64.URLEncoding.EncodeToString,
	}
}

func Fprint(w io.Writer, x interface{}) (err error) {
	return New(w).Print(x)
}

func Fprintln(w io.Writer, x interface{}) (err error) {
	return New(w).Println(x)
}

func Println(x interface{}) (err error) {
	return Fprintln(os.Stdout, x)
}

func Print(x interface{}) error {
	return Fprint(os.Stdout, x)
}

func Sprint(x interface{}) string {
	var buf bytes.Buffer
	Fprint(&buf, x)
	return buf.String()
}

func Sprintln(x interface{}) string {
	var buf bytes.Buffer
	Fprintln(&buf, x)
	return buf.String()
}

type Printer struct {
	Writer      io.Writer
	ByteEncoder func([]byte) string
}

type printerState struct {
	pp        Printer
	indention string
	padding   string
	err       error
}

func (pp Printer) Print(x interface{}) error {
	return pp.print(x, false)
}

func (pp Printer) Println(x interface{}) error {
	return pp.print(x, true)
}

func (pp Printer) print(x interface{}, nl bool) error {
	pps := &printerState{pp: pp}
	xtype := reflect.TypeOf(x)
	if xtype.Kind() == reflect.Struct {
		pps.printf("%s ", xtype.Name())
	}
	if nl {
		pps.printValueLine(reflect.ValueOf(x), 0)
	} else {
		pps.printValue(reflect.ValueOf(x), 0)
	}
	return pps.err
}

func (pps *printerState) failed() bool {
	return pps.err != nil
}

func (pps *printerState) printValueLine(value reflect.Value, n int) {
	pps.printValue(value, n)
	pps.printf("\n")
}

func (pps *printerState) printValue(value reflect.Value, n int) {
	if pps.failed() {
		// short-circuit if an error has been encountered
		return
	}

	vtype := value.Type()

	switch vtype.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			pps.printf("<nil>")
		} else {
			pps.printValue(reflect.Indirect(value), n)
		}
	case reflect.Struct:
		var nfields int
		var longest_name int
		for i := 0; i < vtype.NumField(); i++ {
			name := vtype.Field(i).Name
			if !isExported(name) {
				continue
			}

			nfields++
			if vtype.Field(i).Type.Kind() == reflect.Struct {
				continue
			}
			if nlen := len(vtype.Field(i).Name); nlen > longest_name {
				longest_name = nlen
			}
		}
		if nfields == 0 {
			pps.printf("{}")
			break
		}
		pps.printf("{\n")
		for i := 0; i < vtype.NumField(); i++ {
			name := vtype.Field(i).Name
			if !isExported(name) {
				continue
			}
			if vtype.Field(i).Type.Kind() == reflect.Struct {
				pps.iprintf(n+1, "%s ", name)
			} else {
				pps.iprintf(n+1, "%s %s= ", name,
					pps.pad(longest_name-len(name)))
			}
			pps.printValueLine(value.Field(i), n+1)
		}
		pps.iprintf(n, "}")
	case reflect.Slice, reflect.Array:
		if vtype.Elem().Kind() == reflect.Uint8 {
			pps.printf("%s", pps.encode(value.Interface().([]byte)))
			break
		}
		if value.Len() == 0 {
			pps.printf("[]")
			break
		}
		pps.printf("[\n")
		for i := 0; i < value.Len(); i++ {
			pps.iprintf(n+1, "")
			pps.printValue(value.Index(i), n+1)
			pps.printf(",\n")
		}
		pps.iprintf(n, "]")
	case reflect.String:
		pps.printf("%q", value.Interface())
	case reflect.Map:
		keys := value.MapKeys()
		if len(keys) == 0 {
			pps.printf("{}")
			break
		}
		pps.printf("{\n")
		for _, key := range keys {
			pps.iprintf(n+1, "")
			pps.printValue(key, n+1)
			pps.printf(": ")
			pps.printValue(value.MapIndex(key), n+1)
			pps.printf(",\n")
		}
		pps.iprintf(n, "}")
	case reflect.Interface:
		elem := value.Elem()
		if elem.IsValid() {
			pps.printValue(elem, n)
		} else {
			pps.printf("nil")
		}
	default:
		pps.printf("%v", value.Interface())
	}
}

func (pps *printerState) encode(data []byte) string {
	if pps.pp.ByteEncoder == nil {
		return base64.URLEncoding.EncodeToString(data)
	}
	return pps.pp.ByteEncoder(data)
}

func (pps *printerState) indent(n int) string {
	if len(pps.indention) < n*len(indention) {
		pps.indention = strings.Repeat(indention, max(n, 10))
	}
	return pps.indention[:len(indention)*n]
}

func (pps *printerState) pad(n int) string {
	if len(pps.padding) < n*len(padding) {
		pps.padding = strings.Repeat(padding, max(n, 10))
	}
	return pps.padding[:len(padding)*n]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (pps *printerState) iprintf(indention int, format string,
	args ...interface{}) {

	if pps.failed() {
		return
	}
	_, pps.err = fmt.Fprintf(pps.pp.Writer, "%s"+format,
		append([]interface{}{pps.indent(indention)}, args...)...)
}

func (pps *printerState) printf(format string, args ...interface{}) {
	if pps.failed() {
		return
	}
	_, pps.err = fmt.Fprintf(pps.pp.Writer, format, args...)
}

func isExported(name string) bool {
	if name == "" {
		return false
	}
	return unicode.IsUpper([]rune(name)[0])
}
