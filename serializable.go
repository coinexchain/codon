package codon

import (
	"io"
	"reflect"
	"strings"
)

var serializationTemplate = `
func (ptr *AAA) ToBytes() []byte {
	wBuf := make([]byte, 0, 64)
	EncodeAAA(&wBuf, *ptr)
	return wBuf
}

func (ptr *AAA) FromBytes(bz []byte) {
	var total int
	var err error
	*ptr, total, err = DecodeAAA(bz)
	if total != len(bz) {
		panic("Length Mismatch During Decoding")
	}
	if err != nil {
		panic(err)
	}
}

func (ptr *AAA) DeepCopy() interface{} {
	return DeepCopyAAA(*ptr)
}
`

func GenerateSerializableImpl(
	//output target
	w io.Writer,
	// contains the types which should be regarded as leaf types
	// Key is the full type name, Value is the short type name
	leafTypes map[string]string,
	// Some struct->interface implementation relationship must be ignored
	// Key is struct's alias and Value is interface's alias
	ignoreImpl map[string]string,
	// The types for which we will generate code
	typeEntryList []TypeEntry,
	// extra logics to put in the generated code
	extraLogics string,
	// extra imported packages to put in the generated code
	extraImports []string) {

	// The beginning of the generated file
	w.Write([]byte("//nolint\npackage codec\nimport (\n"))
	for _, p := range extraImports {
		w.Write([]byte(p + "\n"))
	}
	w.Write([]byte("\"encoding/binary\"\n\"errors\"\n)\n"))
	w.Write([]byte(headerLogics))
	w.Write([]byte(extraLogics))

	// Now initialize the context
	ctx := newContext(leafTypes, ignoreImpl)
	for _, entry := range typeEntryList {
		ctx.register(entry.Alias, entry.Name, entry.Value)
	}
	ctx.analyzeIfc()

	// Generate functions for structs
	for _, entry := range typeEntryList {
		t := derefPtr(entry.Value)
		if t.Kind() != reflect.Interface {
			w.Write([]byte("// Non-Interface\n"))
			lines := ctx.generateStructFunc(entry.Alias, t)
			writeLines(w, lines)
			line := strings.Replace(serializationTemplate, "AAA", entry.Alias, -1)
			writeLines(w, []string{line})
		}
	}
}
