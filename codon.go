package codon

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

const (
	MaxFieldNum = 18999
	MaxMagicNum = 536_870_911
	MinMagicNum = 20000
)

func ShowInfoForVar(leafTypes map[string]string, v interface{}) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Print the information header
	fmt.Printf("======= %v '%s' '%s' == \n", t, t.PkgPath(), t.Name())
	showInfo(leafTypes, "", t)
}

func structHasPrivateField(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var isPrivate bool
		for _, r := range field.Name {
			isPrivate = unicode.IsLower(r)
			break
		}
		if isPrivate {
			return true
		}
	}
	return false
}

func showInfo(leafTypes map[string]string, indent string, t reflect.Type) {
	ending := ""
	indentP := indent + "    "
	switch t.Kind() {
	case reflect.Bool:
		fmt.Printf("bool")
	case reflect.Int:
		fmt.Printf("int")
	case reflect.Int8:
		fmt.Printf("int8")
	case reflect.Int16:
		fmt.Printf("int16")
	case reflect.Int32:
		fmt.Printf("int32")
	case reflect.Int64:
		fmt.Printf("int64")
	case reflect.Uint:
		fmt.Printf("uint")
	case reflect.Uint8:
		fmt.Printf("uint8")
	case reflect.Uint16:
		fmt.Printf("uint16")
	case reflect.Uint32:
		fmt.Printf("uint32")
	case reflect.Uint64:
		fmt.Printf("uint64")
	case reflect.Uintptr:
		fmt.Printf("Uintptr!")
	case reflect.Complex64:
		fmt.Printf("complex64!")
	case reflect.Complex128:
		fmt.Printf("complex128!")
	case reflect.Float32:
		fmt.Printf("float32")
	case reflect.Float64:
		fmt.Printf("float64")
	case reflect.Chan:
		fmt.Printf("chan!")
	case reflect.Func:
		fmt.Printf("func!")
	case reflect.Interface:
		fmt.Printf("interface (%s %s)!", t.PkgPath(), t.Name())
	case reflect.Map:
		fmt.Printf("map!")
	case reflect.Ptr:
		path := t.Elem().PkgPath() + "." + t.Elem().Name()
		if _, ok := leafTypes[path]; ok { // Stop when meeting a leaf type
			fmt.Printf("pointer ('%s' '%s')\n", t.Elem().PkgPath(), t.Elem().Name())
		} else {
			fmt.Printf("pointer ('%s' '%s') {\n", t.Elem().PkgPath(), t.Elem().Name())
			fmt.Printf("%s", indentP)
			showInfo(leafTypes, indentP, t.Elem())
			ending = indent + "} // pointer"
		}
	case reflect.Array:
		fmt.Printf("array {\n")
		fmt.Printf("%s", indentP)
		showInfo(leafTypes, indentP, t.Elem())
		ending = indent + "} //array"
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			fmt.Printf("ByteSlice")
		} else {
			fmt.Printf("slice {\n")
			fmt.Printf("%s", indentP)
			showInfo(leafTypes, indentP, t.Elem())
			ending = indent + "} //slice"
		}
	case reflect.String:
		fmt.Printf("string")
	case reflect.Struct:
		path := t.PkgPath() + "." + t.Name()
		if _, ok := leafTypes[path]; ok { // Stop when meeting a leaf type
			fmt.Printf("struct ('%s' '%s')\n", t.PkgPath(), t.Name())
		} else {
			if structHasPrivateField(t) {
				fmt.Printf("struct_with_private {\n")
			} else {
				fmt.Printf("struct {\n")
			}
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				fmt.Printf("%s%s : ('%s' '%s') ", indentP, field.Name, field.Type.PkgPath(), field.Type.Name())
				path = field.Type.PkgPath() + "." + field.Type.Name()
				if _, ok := leafTypes[path]; ok {
					fmt.Printf("\n")
				} else {
					showInfo(leafTypes, indentP, field.Type)
				}
			}
			ending = indent + "} //struct"
		}
	default:
		fmt.Printf("Unknown Kind! %s", t.Kind())
	}

	fmt.Printf("%s\n", ending)
}

func calcMagicNum(lines []string) uint32 {
	h := sha256.New()
	for _, line := range lines {
		h.Write([]byte(line))
	}
	bz := h.Sum(nil)
	val64 := binary.LittleEndian.Uint64(bz[:8])
	val64 = val64 % (MaxMagicNum - MinMagicNum)
	val64 = val64 + MinMagicNum
	if val64 > MaxMagicNum || val64 < MinMagicNum {
		panic("MagicNum calculation error")
	}
	return uint32(val64)
}

type TypeEntry struct {
	Alias string
	Name string
	Value interface{}
}

func writeLines(w io.Writer, lines []string) {
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		w.Write([]byte(line))
		w.Write([]byte("\n"))
	}
}

func GenerateCodecFile(
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
		}
	}
	// Generate functions for interfaces
	for _, entry := range typeEntryList {
		t := derefPtr(entry.Value)
		if t.Kind() == reflect.Interface {
			w.Write([]byte("// Interface\n"))
			lines := ctx.generateIfcFunc(entry.Alias, t)
			writeLines(w, lines)
		}
	}
	// Generate the "getMagicNum" and "getMagicNumOfVar" functions, which maps aliases to magic bytes
	lines := ctx.generateMagicNumFunc()
	writeLines(w, lines)

	// Get sorted list of struct aliases
	aliases := make([]string, 0, len(ctx.structPath2Alias))
	for _, alias := range ctx.structPath2Alias {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	// Top-level encode function, which supports all the registered types. It writes magic bytes at the beginning
	lines = generateIfcEncodeFunc("EncodeAny", aliases)
	writeLines(w, lines)
	// Top-level decode function, which supports all the registered types. It uses magic bytes to decide type
	lines = ctx.generateDecodeAnyFunc()
	writeLines(w, lines)
	// Assign structs to interfaces' pointers
	lines = ctx.generateIfcAssignFunc()
	writeLines(w, lines)
	// Fill an interface object, randomly select the underlying struct type and randomly fill the fields
	lines, aliases = ctx.generateIfcRandFunc("RandAny", "interface{}", nil, aliases, nil)
	writeLines(w, lines)
	// DeepCopy an interface object
	lines = ctx.generateIfcDeepCopyFunc("DeepCopyAny", "interface{}", nil, aliases)
	writeLines(w, lines)
	// Generate a GetSupportList function which returns the sorted full path list of all the supported types
	lines = ctx.generateSupportListFunc()
	writeLines(w, lines)
}

type context struct {
	structPath2Alias map[string]string
	ifcPath2Alias    map[string]string
	structPath2Type  map[string]reflect.Type
	structAlias2Type map[string]reflect.Type
	ifcPath2Type     map[string]reflect.Type

	// map an interface to its implementations
	ifcPath2StructPaths map[string][]string

	structAlias2MagicNum map[string]uint32
	magicNum2StructAlias map[uint32]string

	leafTypes  map[string]string
	ignoreImpl map[string]string
}

func newContext(leafTypes, ignoreImpl map[string]string) *context {
	return &context{
		structPath2Alias: make(map[string]string),
		ifcPath2Alias:    make(map[string]string),
		structPath2Type:  make(map[string]reflect.Type),
		structAlias2Type: make(map[string]reflect.Type),
		ifcPath2Type:     make(map[string]reflect.Type),

		ifcPath2StructPaths:    make(map[string][]string),
		structAlias2MagicNum:   make(map[string]uint32),
		magicNum2StructAlias:   make(map[uint32]string),
		leafTypes:              leafTypes,
		ignoreImpl:             ignoreImpl,
	}
}

func generateIfcEncodeFunc(funcName string, aliases []string) []string {
	lines := make([]string, 0, 1000)
	lines = append(lines, "func "+funcName+"(w *[]byte, x interface{}) {")

	lines = append(lines, "switch v := x.(type) {")
	for _, alias := range aliases {
		lines = append(lines, fmt.Sprintf("case %s:", alias))
		line := fmt.Sprintf("codonEncodeByteSlice(int(getMagicNum(\"%s\")), w, func() []byte {", alias)
		lines = append(lines, line)
		lines = append(lines, "wBuf := make([]byte, 0, 64)\nw := &wBuf")
		lines = append(lines, fmt.Sprintf("Encode%s(w, v)", alias))
		lines = append(lines, "return wBuf\n}())")

		lines = append(lines, fmt.Sprintf("case *%s:", alias))
		line = fmt.Sprintf("codonEncodeByteSlice(int(getMagicNum(\"%s\")), w, func() []byte {", alias)
		lines = append(lines, line)
		lines = append(lines, "wBuf := make([]byte, 0, 64)\nw := &wBuf")
		lines = append(lines, fmt.Sprintf("Encode%s(w, *v)", alias))
		lines = append(lines, "return wBuf\n}())")
	}
	lines = append(lines, "default:")
	lines = append(lines, `panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))`)
	lines = append(lines, "} // end of switch")
	lines = append(lines, "} // end of func")
	return lines
}

func (ctx *context) generateIfcRandFunc(funcName, ifcAlias string, ifcType reflect.Type, aliases []string, ignoreImpl map[string]string) ([]string, []string) {
	lines := make([]string, 0, 1000)
	lines = append(lines, "func "+funcName+"(r RandSrc) "+ifcAlias+" {")
	newAliases := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		if ignoreImpl == nil || ignoreImpl[alias] != ifcAlias {
			newAliases = append(newAliases, alias)
		}
	}
	lines = append(lines, fmt.Sprintf("switch r.GetUint() %% %d {", len(newAliases)))
	for i, alias := range newAliases {
		lines = append(lines, fmt.Sprintf("case %d:", i))
		structType, ok := ctx.structAlias2Type[alias]
		if !ok {
			panic("Can not find type for "+alias)
		}
		if ifcType == nil || structType.Implements(ifcType) {
			lines = append(lines, fmt.Sprintf("return Rand%s(r)", alias))
		} else if reflect.PtrTo(structType).Implements(ifcType) {
			lines = append(lines, fmt.Sprintf("tmp := Rand%s(r)\nreturn &tmp", alias))
		} else {
			panic(alias + "does not implement " + ifcAlias)
		}
	}
	lines = append(lines, "default:")
	lines = append(lines, "panic(\"Unknown Type.\")")
	lines = append(lines, "} // end of switch")
	lines = append(lines, "} // end of func")
	return lines, newAliases
}

func (ctx *context) generateIfcDeepCopyFunc(funcName, ifcAlias string, ifcType reflect.Type, aliases []string) []string {
	lines := make([]string, 0, 1000)
	lines = append(lines, fmt.Sprintf("func %s(x %s) %s {", funcName, ifcAlias, ifcAlias))
	lines = append(lines, "switch v := x.(type) {")
	for _, alias := range aliases {
		structType, ok := ctx.structAlias2Type[alias]
		if !ok {
			panic("Can not find type for "+alias)
		}
		if ifcType == nil || structType.Implements(ifcType) {
			lines = append(lines, fmt.Sprintf("case %s:", alias))
			lines = append(lines, fmt.Sprintf("res := DeepCopy%s(v)\nreturn res", alias))
		}
		if ifcType == nil || reflect.PtrTo(structType).Implements(ifcType) {
			lines = append(lines, fmt.Sprintf("case *%s:", alias))
			lines = append(lines, fmt.Sprintf("res := DeepCopy%s(*v)\nreturn &res", alias))
		}
	}
	lines = append(lines, "default:")
	lines = append(lines, `panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))`)
	lines = append(lines, "} // end of switch")
	lines = append(lines, "} // end of func")
	return lines
}

func (ctx *context) generateDecodeAnyFunc() []string {
	res, _ := ctx.generateIfcDecodeFunc("DecodeAny", "interface{}", nil, ctx.structAlias2MagicNum)
	return res
}

var beforeDecodeFunc = `l := codonDecodeUint64(bz, &n, &err)
if err != nil {
	return
}
bz = bz[n:]
total += n
if int(l) > len(bz) {
	err = errors.New("Length Too Large")
	return
}`

var afterDecodeFunc = `if int(l) != n {
	err = errors.New("Length Mismatch")
	return
}`

var ending = "\nif err != nil {return}\nbz = bz[n:]\ntotal+=n"

func (ctx *context) generateIfcDecodeFunc(funcName, decTypeName string, decType reflect.Type, alias2bytes map[string]uint32) ([]string, []string) {
	aliases := make([]string, 0, len(alias2bytes))
	for alias := range alias2bytes {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	lines := make([]string, 0, 1000)
	lines = append(lines, "func "+funcName+"(bz []byte) (v "+decTypeName+", total int, err error) {")
	lines = append(lines, `
	var n int
	tag := codonDecodeUint64(bz, &n, &err)
	if err != nil {
		return
	}
	bz = bz[n:]
	total += n
	magicNum := uint32(tag >> 3)`)

	lines = append(lines, "switch magicNum {")
	for _, alias := range aliases {
		magicNum := alias2bytes[alias]
		lines = append(lines, fmt.Sprintf("case %d:", magicNum))
		lines = append(lines, beforeDecodeFunc)
		lines = append(lines, fmt.Sprintf("var tmp %s", alias))
		lines = append(lines, fmt.Sprintf("tmp, n, err = Decode%s(bz[:l])%s", alias, ending))
		lines = append(lines, afterDecodeFunc)
		structType, ok := ctx.structAlias2Type[alias]
		if !ok {
			panic("Can not find type for "+alias)
		}
		if decType == nil || structType.Implements(decType) {
			lines = append(lines, "v = tmp\nreturn")
		} else if reflect.PtrTo(structType).Implements(decType) {
			lines = append(lines, "v = &tmp\nreturn")
		} else {
		}
	}
	lines = append(lines, "default:")
	lines = append(lines, "panic(\"Unknown type\")")
	lines = append(lines, "} // end of switch")
	lines = append(lines, "return v, n, nil")
	lines = append(lines, "} // end of "+funcName)
	return lines, aliases
}

func (ctx *context) generateMagicNumFunc() []string {
	lines := make([]string, 0, 1000)
	lines = append(lines, "func getMagicNum(name string) uint32 {")
	lines = append(lines, "switch name {")
	aliases := make([]string, 0, len(ctx.structAlias2MagicNum))
	for alias := range ctx.structAlias2MagicNum {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	for _, alias := range aliases {
		magicNum := ctx.structAlias2MagicNum[alias]
		lines = append(lines, fmt.Sprintf("case \"%s\":", alias))
		lines = append(lines, fmt.Sprintf("return %d", magicNum))
	}
	lines = append(lines, "} // end of switch")
	lines = append(lines, "panic(\"Should not reach here\")")
	lines = append(lines, "return 0")
	lines = append(lines, "} // end of getMagicNum")

	lines = append(lines, "func getMagicNumOfVar(x interface{}) (uint32, bool) {")
	lines = append(lines, "switch x.(type) {")
	for _, alias := range aliases {
		lines = append(lines, fmt.Sprintf("case *%s, %s:", alias, alias))
		magicNum := ctx.structAlias2MagicNum[alias]
		lines = append(lines, fmt.Sprintf("return %d, true", magicNum))
	}
	lines = append(lines, "default:\nreturn 0, false")
	lines = append(lines, "} // end of switch")
	lines = append(lines, "} // end of func")
	return lines
}

func (ctx *context) generateSupportListFunc() []string {
	length := len(ctx.structPath2Alias) + len(ctx.ifcPath2Alias) + 10
	paths := make([]string, 0, length)
	for path := range ctx.structPath2Alias {
		paths = append(paths, path)
	}
	for path := range ctx.ifcPath2Alias {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	lines := make([]string, 0, length)
	lines = append(lines, "func GetSupportList() []string {")
	lines = append(lines, "return []string {")
	for _, path := range paths {
		lines = append(lines, fmt.Sprintf("\"%s\",", path))
	}
	lines = append(lines, "}")
	lines = append(lines, "} // end of GetSupportList")
	return lines
}

func (ctx *context) analyzeIfc() {
	for ifcPath, ifcType := range ctx.ifcPath2Type {
		for structPath, structType := range ctx.structPath2Type {
			if structType.Implements(ifcType) || reflect.PtrTo(structType).Implements(ifcType) {
				if _, ok := ctx.ifcPath2StructPaths[ifcPath]; ok {
					ctx.ifcPath2StructPaths[ifcPath] = append(ctx.ifcPath2StructPaths[ifcPath], structPath)
				} else {
					ctx.ifcPath2StructPaths[ifcPath] = []string{structPath}
				}
			}
		}
	}
}

func derefPtr(v interface{}) reflect.Type {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (ctx *context) register(alias string, name string, v interface{}) {
	t := derefPtr(v)
	path := t.PkgPath() + "." + t.Name()
	//if len(t.PkgPath()) == 0 || len(t.Name()) == 0 {
	//	panic("Invalid Path:" + path)
	//}
	if t.Kind() == reflect.Interface {
		ctx.ifcPath2Alias[path] = alias
		ctx.ifcPath2Type[path] = t
	} else {
		if path == "." { // a special case for [][]byte
			ctx.structPath2Alias[alias] = alias
		} else {
			ctx.structPath2Alias[path] = alias
		}
		ctx.structPath2Type[path] = t
		ctx.structAlias2Type[alias] = t
		magicNum := calcMagicNum([]string{alias, name})
		if otherAlias, ok := ctx.magicNum2StructAlias[magicNum]; ok {
			panic("Magic Bytes Conflicts: " + otherAlias + " vs " + alias)
		}
		ctx.structAlias2MagicNum[alias] = magicNum
		ctx.magicNum2StructAlias[magicNum] = alias
	}
}

func (ctx *context) generateIfcAssignFunc() []string {
	lines := make([]string, 0, 1000)
	lines = append(lines, "func AssignIfcPtrFromStruct(ifcPtrIn interface{}, structObjIn interface{}) {")
	lines = append(lines, "switch ifcPtr := ifcPtrIn.(type) {")
	for ifcPath, structPaths := range ctx.ifcPath2StructPaths {
		ifcAlias, ok := ctx.ifcPath2Alias[ifcPath]
		if !ok {
			panic("cannot find ifcAlias")
		}
		lines = append(lines, fmt.Sprintf("case *%s:", ifcAlias))
		lines = append(lines, "switch structObj := structObjIn.(type) {")
		for _, structPath := range structPaths {
			structAlias, ok := ctx.structPath2Alias[structPath]
			if !ok {
				panic("cannot find structAlias")
			}
			lines = append(lines, fmt.Sprintf("\tcase %s:", structAlias))
			lines = append(lines, "\t*ifcPtr = &structObj")
		}
		lines = append(lines, "\tdefault:")
		lines = append(lines, `	panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))`)
		lines = append(lines, "\t} // end switch of structs")
	}
	lines = append(lines, "default:")
	lines = append(lines, `panic(fmt.Sprintf("Unknown Type %v\n", reflect.TypeOf(ifcPtrIn)))`)
	lines = append(lines, "} // end switch of interfaces")
	lines = append(lines, "}")
	return lines
}

func (ctx *context) generateIfcFunc(ifc string, t reflect.Type) []string {
	ifcPath := t.PkgPath() + "." + t.Name()
	structPaths, ok := ctx.ifcPath2StructPaths[ifcPath]
	if !ok {
		panic("Cannot find implementations for " + ifc)
	}
	alias2bytes := make(map[string]uint32, len(structPaths))
	for _, structPath := range structPaths {
		alias, ok := ctx.structPath2Alias[structPath]
		if !ok {
			panic("Cannot find alias")
		}
		magicNum, ok := ctx.structAlias2MagicNum[alias]
		if !ok {
			panic("Cannot find magicbytes")
		}
		alias2bytes[alias] = magicNum
	}
	decLines, aliases := ctx.generateIfcDecodeFunc("Decode"+ifc, ifc, t, alias2bytes)
	encLines := generateIfcEncodeFunc("Encode"+ifc, aliases)
	randLines, aliases := ctx.generateIfcRandFunc("Rand"+ifc, ifc, t, aliases, ctx.ignoreImpl)
	deepcopyLines := ctx.generateIfcDeepCopyFunc("DeepCopy"+ifc, ifc, t, aliases)
	result := make([]string, 0, len(decLines)+len(encLines)+len(randLines)+len(deepcopyLines))
	result = append(result, decLines...)
	result = append(result, encLines...)
	result = append(result, randLines...)
	result = append(result, deepcopyLines...)
	return result
}

func (ctx *context) generateStructFunc(alias string, t reflect.Type) []string {
	lines := make([]string, 0, 1000)

	// Encode
	line := fmt.Sprintf("func Encode%s(w *[]byte, v %s) {", alias, alias)
	lines = append(lines, line)
	_, isLeaf := ctx.leafTypes[t.PkgPath()+"."+t.Name()]
	if !isLeaf && len(t.PkgPath())==0 {
		isLeaf = true
	}
	if t.Kind() == reflect.Struct && !isLeaf{
		ctx.genStructEncLines(t, &lines, "v", 0)
	} else {
		ctx.genFieldEncLines(0, t, &lines, "v", 0)
	}
	lines = append(lines, "} //End of Encode"+alias+"\n")

	// Decode
	line = fmt.Sprintf("func Decode%s(bz []byte) (v %s, total int, err error) {", alias, alias)
	lines = append(lines, line)
	lines = append(lines, "var n int")
	ending := "\nif err != nil {return v, total, err}\nbz = bz[n:]\ntotal+=n"
	lines = append(lines, "for len(bz) != 0 {")
	lines = append(lines, fmt.Sprintf("tag := codonDecodeUint64(bz, &n, &err)%s", ending))
	lines = append(lines, "tag = tag >> 3")
	lines = append(lines, "switch tag {")
	if t.Kind() == reflect.Struct && !isLeaf {
		ctx.genStructDecLines(t, &lines, "v", 0)
	} else {
		lines = append(lines, "case 0:")
		ctx.genFieldDecLines(0, t, &lines, "v", 0)
	}
	lines = append(lines, "default: err = errors.New(\"Unknown Field\")\nreturn\n}")
	lines = append(lines, "} // end for")
	lines = append(lines, "return v, total, nil")
	lines = append(lines, "} //End of Decode"+alias+"\n")

	// Rand
	line = fmt.Sprintf("func Rand%s(r RandSrc) %s {", alias, alias)
	lines = append(lines, line)
	lengthLinePosition := len(lines)
	lines = append(lines, "") // length placeholder
	lines = append(lines, "var v "+alias)
	needLength := false
	if t.Kind() == reflect.Struct && !isLeaf {
		nl := ctx.genStructRandLines(t, &lines, "v", 0)
		needLength = needLength || nl
	} else {
		nl := ctx.genFieldRandLines(t, &lines, "v", 0)
		needLength = needLength || nl
	}
	if needLength {
		lines[lengthLinePosition] = "var length int"
	}
	lines = append(lines, "return v")
	lines = append(lines, "} //End of Rand"+alias+"\n")

	// DeepCopy
	line = fmt.Sprintf("func DeepCopy%s(in %s) (out %s) {", alias, alias, alias)
	lines = append(lines, line)
	lengthLinePosition = len(lines)
	lines = append(lines, "") // length placeholder
	needLength = false
	if t.Kind() == reflect.Struct && !isLeaf {
		nl := ctx.genStructDeepCopyLines(t, &lines, "", 0)
		needLength = needLength || nl
	} else {
		nl := ctx.genFieldDeepCopyLines(t, &lines, "", 0)
		needLength = needLength || nl
	}
	if needLength {
		lines[lengthLinePosition] = "var length int"
	}
	lines = append(lines, "return")
	lines = append(lines, "} //End of DeepCopy"+alias+"\n")

	return lines
}

//====================================================================

func isMutex(t reflect.Type) bool {
	if t.PkgPath() == "sync" {
		if t.Name() == "Mutex" || t.Name() == "RWMutex" {
			return true
		}
	}
	return false
}

func (ctx *context) genFieldEncLines(fieldNum int, t reflect.Type, lines *[]string, fieldName string, iterLevel int) {
	if fieldNum > MaxFieldNum {
		panic("Field Number is too large")
	}
	if isMutex(t) {
		return
	}
	isPtr := false
	if t.Kind() == reflect.Ptr {
		isPtr = true
		elemT := t.Elem()
		if elemT.Kind() == reflect.Struct {
			t = elemT
		} else {
			panic(fmt.Sprintf("Pointer to %s is not supported", elemT.Kind()))
		}
	}
	var line string
	switch t.Kind() {
	case reflect.Chan:
		panic("Channel is not supported")
	case reflect.Func:
		panic("Func is not supported")
	case reflect.Uintptr:
		panic("Uintptr is not supported")
	case reflect.Complex64:
		panic("Complex64 is not supported")
	case reflect.Complex128:
		panic("Complex128 is not supported")
	case reflect.Map:
		panic("Map is not supported")
	case reflect.Float32:
		panic("Float32 is not supported")
	case reflect.Float64:
		panic("Float64 is not supported")

	case reflect.Bool:
		if len(t.PkgPath()) == 0 {
			line = fmt.Sprintf("codonEncodeBool(%d, w, %s)", fieldNum, fieldName)
		} else {
			line = fmt.Sprintf("codonEncodeBool(%d, w, bool(%s))", fieldNum, fieldName)
		}
	case reflect.Int:
		line = fmt.Sprintf("codonEncodeVarint(%d, w, int64(%s))", fieldNum, fieldName)
	case reflect.Int8:
		if len(t.PkgPath()) == 0 {
			line = fmt.Sprintf("codonEncodeInt8(%d, w, %s)", fieldNum, fieldName)
		} else {
			line = fmt.Sprintf("codonEncodeInt8(%d, w, int8(%s))", fieldNum, fieldName)
		}
	case reflect.Int16:
		if len(t.PkgPath()) == 0 {
			line = fmt.Sprintf("codonEncodeInt16(%d, w, %s)", fieldNum, fieldName)
		} else {
			line = fmt.Sprintf("codonEncodeInt16(%d, w, int16(%s))", fieldNum, fieldName)
		}
	case reflect.Int32:
		line = fmt.Sprintf("codonEncodeVarint(%d, w, int64(%s))", fieldNum, fieldName)
	case reflect.Int64:
		line = fmt.Sprintf("codonEncodeVarint(%d, w, int64(%s))", fieldNum, fieldName)
	case reflect.Uint:
		line = fmt.Sprintf("codonEncodeUvarint(%d, w, uint64(%s))", fieldNum, fieldName)
	case reflect.Uint8:
		if len(t.PkgPath()) == 0 {
			line = fmt.Sprintf("codonEncodeUint8(%d, w, %s)", fieldNum, fieldName)
		} else {
			line = fmt.Sprintf("codonEncodeUint8(%d, w, uint8(%s))", fieldNum, fieldName)
		}
	case reflect.Uint16:
		if len(t.PkgPath()) == 0 {
			line = fmt.Sprintf("codonEncodeUint16(%d, w, %s)", fieldNum, fieldName)
		} else {
			line = fmt.Sprintf("codonEncodeUint16(%d, w, uint16(%s))", fieldNum, fieldName)
		}
	case reflect.Uint32:
		line = fmt.Sprintf("codonEncodeUvarint(%d, w, uint64(%s))", fieldNum, fieldName)
	case reflect.Uint64:
		line = fmt.Sprintf("codonEncodeUvarint(%d, w, uint64(%s))", fieldNum, fieldName)
	case reflect.String:
		if len(t.PkgPath()) == 0 {
			line = fmt.Sprintf("codonEncodeString(%d, w, %s)", fieldNum, fieldName)
		} else {
			line = fmt.Sprintf("codonEncodeString(%d, w, string(%s))", fieldNum, fieldName)
		}
	case reflect.Array:
		if t.Elem().Kind() != reflect.Uint8 {
			panic("ByteArray is the only supported array type")
		}
		line = fmt.Sprintf("codonEncodeByteSlice(%d, w, %s[:])", fieldNum, fieldName)
	case reflect.Slice:
		elemT := t.Elem()
		if elemT.Kind() == reflect.Uint8 {
			line = fmt.Sprintf("codonEncodeByteSlice(%d, w, %s[:])", fieldNum, fieldName)
		} else {
			*lines = append(*lines, line)
			iterVar := fmt.Sprintf("_%d", iterLevel)
			line = fmt.Sprintf("for %s:=0; %s<len(%s); %s++ {",
				iterVar, iterVar, fieldName, iterVar)
			*lines = append(*lines, line)
			varName := fieldName + "[" + iterVar + "]"
			ctx.genFieldEncLines(fieldNum, elemT, lines, varName, iterLevel+1)
			line = "}"
		}
	case reflect.Interface:
		typePath := t.PkgPath() + "." + t.Name()
		alias, ok := ctx.ifcPath2Alias[typePath]
		if !ok {
			panic("Cannot find alias for:" + typePath)
		}
		line = fmt.Sprintf("codonEncodeByteSlice(%d, w, func() []byte {", fieldNum)
		*lines = append(*lines, line)
		*lines = append(*lines, "w := make([]byte, 0, 64)")
		*lines = append(*lines, fmt.Sprintf("Encode%s(&w, %s)// interface_encode", alias, fieldName))
		line = "return w\n}()) // end of " + fieldName
	case reflect.Ptr:
		panic("Should not reach here")
	case reflect.Struct:
		if _, ok := ctx.leafTypes[t.PkgPath()+"."+t.Name()]; ok {
			if isPtr {
				fieldName = "*(" + fieldName + ")"
			}
			line = fmt.Sprintf("codonEncodeByteSlice(%d, w, Encode%s(%s))", fieldNum, t.Name(), fieldName)
		} else {
			line = fmt.Sprintf("codonEncodeByteSlice(%d, w, func() []byte {",fieldNum)
			*lines = append(*lines, line)
			*lines = append(*lines, "wBuf := make([]byte, 0, 64)\nw := &wBuf")
			ctx.genStructEncLines(t, lines, fieldName, iterLevel)
			line = "return wBuf\n}()) // end of " + fieldName
		}
	default:
		panic(fmt.Sprintf("Unknown Kind %s", t.Kind()))
	}
	*lines = append(*lines, line)
}

func (ctx *context) genStructEncLines(t reflect.Type, lines *[]string, varName string, iterLevel int) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		ctx.genFieldEncLines(i+1, field.Type, lines, varName+"."+field.Name, iterLevel)
	}
}

//=========================

func (ctx *context) getTypeName(elemT reflect.Type) string {
	if elemT.Kind() == reflect.Ptr {
		panic("Should not reach here")
	}
	if len(elemT.PkgPath()) == 0 {
		return elemT.Name() //basic type
	}
	typePath := elemT.PkgPath() + "." + elemT.Name()
	alias, ok := ctx.structPath2Alias[typePath]
	if !ok {
		alias, ok = ctx.ifcPath2Alias[typePath]
	}
	if !ok {
		panic(typePath + " is not registered")
	}
	return alias
}

func (ctx *context) getTypeInfo(elemT reflect.Type) (string, bool) {
	isPtr := false
	if elemT.Kind() == reflect.Ptr {
		elemT = elemT.Elem()
		isPtr = true
	}
	if len(elemT.PkgPath()) == 0 {
		return elemT.Name(), isPtr //basic type
	}
	typePath := elemT.PkgPath() + "." + elemT.Name()
	alias, ok := ctx.structPath2Alias[typePath]
	if !ok {
		alias, ok = ctx.ifcPath2Alias[typePath]
	}
	if !ok {
		panic(typePath + " is not registered")
	}
	if !ok {
		if elemT.Kind() == reflect.Array {
			panic("Do not support slice/array of arrays")
		}
		if elemT.Kind() == reflect.Slice {
			if elemT.Elem().Kind() == reflect.Uint8 {
				return "[]byte", false
			} else {
				panic("Do not support slice/array of slices, except for byte slices")
			}
		}
	}
	return alias, isPtr
}

func (ctx *context) buildDecLine(typeName, fieldName, ending string, t reflect.Type) string {
	if len(t.PkgPath()) == 0 {
		return fmt.Sprintf("%s = %s(codonDecode%s(bz, &n, &err))%s", fieldName, strings.ToLower(typeName), typeName, ending)
	}
	alias := ctx.getTypeName(t)
	return fmt.Sprintf("%s = %s(codonDecode%s(bz, &n, &err))%s", fieldName, alias, typeName, ending)
}

func (ctx *context) initPtrMember(fieldName string, t reflect.Type) string {
	typePath := t.PkgPath() + "." + t.Name()
	alias, ok := ctx.structPath2Alias[typePath]
	if !ok {
		alias, ok = ctx.leafTypes[typePath]
	}
	if !ok {
		panic("Cannot find alias for:" + typePath)
	}
	return fmt.Sprintf("%s = &%s{}", fieldName, alias)
}

func (ctx *context) genFieldDecLines(fieldNum int, t reflect.Type, lines *[]string, fieldName string, iterLevel int) {
	if fieldNum >= MaxFieldNum {
		panic("Field Number is too large")
	}
	if isMutex(t) {
		return
	}
	isPtr := false
	if t.Kind() == reflect.Ptr {
		isPtr = true
		elemT := t.Elem()
		if elemT.Kind() == reflect.Struct {
			t = elemT
		} else {
			panic(fmt.Sprintf("Pointer to %s is not supported", elemT.Kind()))
		}
	}
	var line string
	switch t.Kind() {
	case reflect.Chan:
		panic("Channel is not supported")
	case reflect.Func:
		panic("Func is not supported")
	case reflect.Uintptr:
		panic("Uintptr is not supported")
	case reflect.Complex64:
		panic("Complex64 is not supported")
	case reflect.Complex128:
		panic("Complex128 is not supported")
	case reflect.Map:
		panic("Map is not supported")
	case reflect.Float32:
		panic("Float32 is not supported")
	case reflect.Float64:
		panic("Float64 is not supported")
	case reflect.Bool:
		line = ctx.buildDecLine("Bool", fieldName, ending, t)
	case reflect.Int:
		line = ctx.buildDecLine("Int", fieldName, ending, t)
	case reflect.Int8:
		line = ctx.buildDecLine("Int8", fieldName, ending, t)
	case reflect.Int16:
		line = ctx.buildDecLine("Int16", fieldName, ending, t)
	case reflect.Int32:
		line = ctx.buildDecLine("Int32", fieldName, ending, t)
	case reflect.Int64:
		line = ctx.buildDecLine("Int64", fieldName, ending, t)
	case reflect.Uint:
		line = ctx.buildDecLine("Uint", fieldName, ending, t)
	case reflect.Uint8:
		line = ctx.buildDecLine("Uint8", fieldName, ending, t)
	case reflect.Uint16:
		line = ctx.buildDecLine("Uint16", fieldName, ending, t)
	case reflect.Uint32:
		line = ctx.buildDecLine("Uint32", fieldName, ending, t)
	case reflect.Uint64:
		line = ctx.buildDecLine("Uint64", fieldName, ending, t)
	case reflect.String:
		line = ctx.buildDecLine("String", fieldName, ending, t)
	case reflect.Array:
		if t.Elem().Kind() != reflect.Uint8 {
			panic("ByteArray is the only supported array type")
		}
		*lines = append(*lines, fmt.Sprintf("o := %s[:]", fieldName))
		line = fmt.Sprintf("n, err = codonGetByteSlice(&o, bz)%s", ending)
	case reflect.Slice:
		typeName, isPtr := ctx.getTypeInfo(t.Elem())
		elemT := t.Elem()
		if isPtr {
			if elemT.Kind() == reflect.Interface || elemT.Kind() == reflect.Struct {
				*lines = append(*lines, beforeDecodeFunc)
				line = fmt.Sprintf("var tmp %s\ntmp, n, err = Decode%s(bz[:l])%s",
					typeName, typeName, ending)
				*lines = append(*lines, line)
				*lines = append(*lines, afterDecodeFunc)
			} else {
				ctx.genFieldDecLines(fieldNum, elemT, lines, "tmp", iterLevel+1)
			}
			line = fmt.Sprintf("%s = append(%s, &tmp)", fieldName, fieldName)
		} else {
			if elemT.Kind() == reflect.Uint8 {
				line = fmt.Sprintf("var tmpBz []byte\nn, err = codonGetByteSlice(&tmpBz, bz)%s", ending)
				*lines = append(*lines, line)
				line = fmt.Sprintf("%s = tmpBz", fieldName)
			} else {
				if elemT.Kind() == reflect.Interface || elemT.Kind() == reflect.Struct {
					*lines = append(*lines, beforeDecodeFunc)
					line = fmt.Sprintf("var tmp %s\ntmp, n, err = Decode%s(bz[:l])%s",
						typeName, typeName, ending)
					*lines = append(*lines, line)
					*lines = append(*lines, afterDecodeFunc)
				} else {
					*lines = append(*lines, fmt.Sprintf("var tmp %s", typeName))
					ctx.genFieldDecLines(fieldNum, elemT, lines, "tmp", iterLevel+1)
				}
				line = fmt.Sprintf("%s = append(%s, tmp)", fieldName, fieldName)
			}
		}
	case reflect.Interface:
		typePath := t.PkgPath() + "." + t.Name()
		alias, ok := ctx.ifcPath2Alias[typePath]
		if !ok {
			panic("Cannot find alias for:" + typePath)
		}
		*lines = append(*lines, beforeDecodeFunc)
		line = fmt.Sprintf("%s, n, err = Decode%s(bz[:l])%s // interface_decode", fieldName, alias, ending)
		*lines = append(*lines, line)
		line = afterDecodeFunc
	case reflect.Ptr:
		panic("Should not reach here")
	case reflect.Struct:
		if _, ok := ctx.leafTypes[t.PkgPath()+"."+t.Name()]; ok {
			*lines = append(*lines, beforeDecodeFunc)
			if isPtr {
				*lines = append(*lines, ctx.initPtrMember(fieldName, t))
				line = fmt.Sprintf("*(%s), n, err = Decode%s(bz[:l])%s", fieldName, t.Name(), ending)
				*lines = append(*lines, line)
			} else {
				line = fmt.Sprintf("%s, n, err = Decode%s(bz[:l])%s", fieldName, t.Name(), ending)
				*lines = append(*lines, line)
			}
			line = afterDecodeFunc
		} else {
			if isPtr {
				*lines = append(*lines, ctx.initPtrMember(fieldName, t))
			}
			*lines = append(*lines, fmt.Sprintf("l := codonDecodeUint64(bz, &n, &err)%s", ending))
			*lines = append(*lines, "func(bz []byte) {")
			*lines = append(*lines, "for len(bz) != 0 {")
			*lines = append(*lines, fmt.Sprintf("tag := codonDecodeUint64(bz, &n, &err)%s", ending))
			*lines = append(*lines, "tag = tag >> 3")
			*lines = append(*lines, "switch tag {")
			ctx.genStructDecLines(t, lines, fieldName, iterLevel)
			*lines = append(*lines, "default: err = errors.New(\"Unknown Field\")\nreturn\n}")
			*lines = append(*lines, "} // end for")
			*lines = append(*lines, "}(bz[:l]) // end func")
			*lines = append(*lines, "if err != nil {return}")
			*lines = append(*lines, "bz = bz[l:]\nn += int(l)")
		}
	default:
		panic(fmt.Sprintf("Unknown Kind %s", t.Kind()))
	}
	*lines = append(*lines, line)
}

func (ctx *context) genStructDecLines(t reflect.Type, lines *[]string, varName string, iterLevel int) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := varName+"."+field.Name
		*lines = append(*lines, fmt.Sprintf("case %d: // %s", i, fieldName))
		ctx.genFieldDecLines(i+1, field.Type, lines, fieldName, iterLevel)
	}
}

//======================

func (ctx *context) buildRandLine(typeName, fieldName string, t reflect.Type) string {
	if len(t.PkgPath()) == 0 {
		return fmt.Sprintf("%s = r.Get%s()", fieldName, typeName)
	}
	alias := ctx.getTypeName(t)
	return fmt.Sprintf("%s = %s(r.Get%s())", fieldName, alias, typeName)
}

func (ctx *context) genFieldRandLines(t reflect.Type, lines *[]string, fieldName string, iterLevel int) bool {
	if isMutex(t) {
		return false
	}
	needLength := false
	isPtr := false
	if t.Kind() == reflect.Ptr {
		isPtr = true
		elemT := t.Elem()
		if elemT.Kind() == reflect.Struct {
			t = elemT
		} else {
			panic(fmt.Sprintf("Pointer to %s is not supported", elemT.Kind()))
		}
	}
	var line string
	switch t.Kind() {
	case reflect.Chan:
		panic("Channel is not supported")
	case reflect.Func:
		panic("Func is not supported")
	case reflect.Uintptr:
		panic("Uintptr is not supported")
	case reflect.Complex64:
		panic("Complex64 is not supported")
	case reflect.Complex128:
		panic("Complex128 is not supported")
	case reflect.Map:
		panic("Map is not supported")
	case reflect.Bool:
		line = ctx.buildRandLine("Bool", fieldName, t)
	case reflect.Int:
		line = ctx.buildRandLine("Int", fieldName, t)
	case reflect.Int8:
		line = ctx.buildRandLine("Int8", fieldName, t)
	case reflect.Int16:
		line = ctx.buildRandLine("Int16", fieldName, t)
	case reflect.Int32:
		line = ctx.buildRandLine("Int32", fieldName, t)
	case reflect.Int64:
		line = ctx.buildRandLine("Int64", fieldName, t)
	case reflect.Uint:
		line = ctx.buildRandLine("Uint", fieldName, t)
	case reflect.Uint8:
		line = ctx.buildRandLine("Uint8", fieldName, t)
	case reflect.Uint16:
		line = ctx.buildRandLine("Uint16", fieldName, t)
	case reflect.Uint32:
		line = ctx.buildRandLine("Uint32", fieldName, t)
	case reflect.Uint64:
		line = ctx.buildRandLine("Uint64", fieldName, t)
	case reflect.Float32:
		line = ctx.buildRandLine("Float32", fieldName, t)
	case reflect.Float64:
		line = ctx.buildRandLine("Float64", fieldName, t)
	case reflect.String:
		if len(t.PkgPath()) == 0 {
			line = fmt.Sprintf("%s = r.GetString(1+int(r.GetUint()%%(MaxStringLength-1)))", fieldName)
		} else {
			alias := ctx.getTypeName(t)
			line = fmt.Sprintf("%s = %s(r.GetString(1+int(r.GetUint()%%(MaxStringLength-1))))", fieldName, alias)
		}
	case reflect.Array, reflect.Slice:
		line = "length = 1+int(r.GetUint()%(MaxSliceLength-1))"
		if t.Kind() == reflect.Array {
			line = fmt.Sprintf("length = %d", t.Len())
		}
		needLength = true
		*lines = append(*lines, line)
		typeName, isPtr := ctx.getTypeInfo(t.Elem())
		elemT := t.Elem()
		if isPtr {
			if t.Kind() == reflect.Slice {
				makeSlice := fmt.Sprintf("if length==0 {%s = nil\n} else {\n%s = make([]*%s, length)\n}",
					fieldName, fieldName, typeName)
				*lines = append(*lines, makeSlice)
			}
			iterVar := fmt.Sprintf("_%d", iterLevel)
			initVar := fmt.Sprintf("%s, length_%d := 0, length", iterVar, iterLevel)
			line = fmt.Sprintf("for %s; %s<length_%d; %s++ { //%s of %s",
				initVar, iterVar, iterLevel, iterVar, t.Kind(), elemT.Kind())
			*lines = append(*lines, line)
			if elemT.Kind() == reflect.Interface || elemT.Kind() == reflect.Struct {
				line = fmt.Sprintf("tmp := Rand%s(bz)%s", typeName)
				*lines = append(*lines, line)
				line = fmt.Sprintf("%s[%s] = &tmp", fieldName, iterVar)
				*lines = append(*lines, line)
			} else {
				varName := fieldName + "[" + iterVar + "]"
				nl := ctx.genFieldRandLines(elemT, lines, varName, iterLevel+1)
				needLength = needLength || nl
			}
			line = "}"
		} else {
			if t.Kind() == reflect.Slice && elemT.Kind() != reflect.Uint8 {
				makeSlice := fmt.Sprintf("if length==0 {%s = nil\n} else {\n%s = make([]%s, length)\n}",
					fieldName, fieldName, typeName)
				*lines = append(*lines, makeSlice)
			}
			if t.Kind() == reflect.Slice && elemT.Kind() == reflect.Uint8 {
				line = fmt.Sprintf("%s = r.GetBytes(length)", fieldName)
			} else {
				iterVar := fmt.Sprintf("_%d", iterLevel)
				initVar := fmt.Sprintf("%s, length_%d := 0, length", iterVar, iterLevel)
				line = fmt.Sprintf("for %s; %s<length_%d; %s++ { //%s of %s",
					initVar, iterVar, iterLevel, iterVar, t.Kind(), elemT.Kind())
				*lines = append(*lines, line)
				if elemT.Kind() == reflect.Interface || elemT.Kind() == reflect.Struct {
					line = fmt.Sprintf("%s[%s] = Rand%s(r)", fieldName, iterVar, typeName)
					*lines = append(*lines, line)
				} else {
					varName := fieldName + "[" + iterVar + "]"
					nl := ctx.genFieldRandLines(elemT, lines, varName, iterLevel+1)
					needLength = needLength || nl
				}
				line = "}"
			}
		}
	case reflect.Interface:
		typePath := t.PkgPath() + "." + t.Name()
		alias, ok := ctx.ifcPath2Alias[typePath]
		if !ok {
			panic("Cannot find alias for:" + typePath)
		}
		line = fmt.Sprintf("%s = Rand%s(r) // interface_decode", fieldName, alias)
	case reflect.Ptr:
		panic("Should not reach here")
	case reflect.Struct:
		if _, ok := ctx.leafTypes[t.PkgPath()+"."+t.Name()]; ok {
			if isPtr {
				*lines = append(*lines, ctx.initPtrMember(fieldName, t))
				line = fmt.Sprintf("*(%s) = Rand%s(r)", fieldName, t.Name())
			} else {
				line = fmt.Sprintf("%s = Rand%s(r)", fieldName, t.Name())
			}
		} else {
			if isPtr {
				*lines = append(*lines, ctx.initPtrMember(fieldName, t))
			}
			nl := ctx.genStructRandLines(t, lines, fieldName, iterLevel)
			needLength = needLength || nl
			line = "// end of " + fieldName
		}
	default:
		panic(fmt.Sprintf("Unknown Kind %s", t.Kind()))
	}
	*lines = append(*lines, line)
	return needLength
}

func (ctx *context) genStructRandLines(t reflect.Type, lines *[]string, varName string, iterLevel int) bool {
	needLength := false
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		nl := ctx.genFieldRandLines(field.Type, lines, varName+"."+field.Name, iterLevel)
		needLength = needLength || nl
	}
	return needLength
}

//===================================================================

func (ctx *context) genFieldDeepCopyLines(t reflect.Type, lines *[]string, fieldName string, iterLevel int) bool {
	if isMutex(t) {
		return false
	}
	needLength := false
	isPtr := false
	if t.Kind() == reflect.Ptr {
		isPtr = true
		elemT := t.Elem()
		if elemT.Kind() == reflect.Struct {
			t = elemT
		} else {
			panic(fmt.Sprintf("Pointer to %s is not supported", elemT.Kind()))
		}
	}
	var line string
	switch t.Kind() {
	case reflect.Chan:
		panic("Channel is not supported")
	case reflect.Func:
		panic("Func is not supported")
	case reflect.Uintptr:
		panic("Uintptr is not supported")
	case reflect.Complex64:
		panic("Complex64 is not supported")
	case reflect.Complex128:
		panic("Complex128 is not supported")
	case reflect.Map:
		panic("Map is not supported")
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
	reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
	reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32,
	reflect.Float64, reflect.String:
		line = fmt.Sprintf("out%s = in%s", fieldName, fieldName)
	case reflect.Array, reflect.Slice:
		line = fmt.Sprintf("length = len(in%s)", fieldName)
		needLength = true
		*lines = append(*lines, line)
		typeName, isPtr := ctx.getTypeInfo(t.Elem())
		elemT := t.Elem()
		if isPtr {
			if t.Kind() == reflect.Slice {
				makeSlice := fmt.Sprintf("if length==0 {out%s = nil\n} else {\nout%s = make([]*%s, length)\n}",
					fieldName, fieldName, typeName)
				*lines = append(*lines, makeSlice)
			}
			iterVar := fmt.Sprintf("_%d", iterLevel)
			initVar := fmt.Sprintf("%s, length_%d := 0, length", iterVar, iterLevel)
			line = fmt.Sprintf("for %s; %s<length_%d; %s++ { //%s of %s",
				initVar, iterVar, iterLevel, iterVar, t.Kind(), elemT.Kind())
			*lines = append(*lines, line)
			if elemT.Kind() == reflect.Interface || elemT.Kind() == reflect.Struct {
				line = fmt.Sprintf("tmp = DeepCopy%s(in%s[%s])", typeName, fieldName, iterVar)
				*lines = append(*lines, line)
				line = fmt.Sprintf("out%s[%s] = &tmp", fieldName, iterVar)
				*lines = append(*lines, line)
			} else {
				varName := fieldName + "[" + iterVar + "]"
				nl := ctx.genFieldDeepCopyLines(elemT, lines, varName, iterLevel+1)
				needLength = needLength || nl
			}
			line = "}"
		} else {
			if t.Kind() == reflect.Slice {
				makeSlice := fmt.Sprintf("if length==0 {out%s = nil\n} else {\nout%s = make([]%s, length)\n}",
					fieldName, fieldName, typeName)
				*lines = append(*lines, makeSlice)
			}
			if elemT.Kind() == reflect.Uint8 && t.Kind() == reflect.Slice {
				line = fmt.Sprintf("copy(out%s[:], in%s[:])", fieldName, fieldName)
			} else {
				iterVar := fmt.Sprintf("_%d", iterLevel)
				initVar := fmt.Sprintf("%s, length_%d := 0, length", iterVar, iterLevel)
				line = fmt.Sprintf("for %s; %s<length_%d; %s++ { //%s of %s",
					initVar, iterVar, iterLevel, iterVar, t.Kind(), t.Elem().Kind())
				*lines = append(*lines, line)
				if t.Elem().Kind() == reflect.Interface || t.Elem().Kind() == reflect.Struct {
					line = fmt.Sprintf("out%s[%s] = DeepCopy%s(in%s[%s])", fieldName, iterVar, typeName, fieldName, iterVar)
					*lines = append(*lines, line)
				} else {
					varName := fieldName + "[" + iterVar + "]"
					nl := ctx.genFieldDeepCopyLines(elemT, lines, varName, iterLevel+1)
					needLength = needLength || nl
				}
				line = "}"
			}
		}
	case reflect.Interface:
		typePath := t.PkgPath() + "." + t.Name()
		alias, ok := ctx.ifcPath2Alias[typePath]
		if !ok {
			panic("Cannot find alias for:" + typePath)
		}
		line = fmt.Sprintf("out%s = DeepCopy%s(in%s)", fieldName, alias, fieldName)
	case reflect.Ptr:
		panic("Should not reach here")
	case reflect.Struct:
		if _, ok := ctx.leafTypes[t.PkgPath()+"."+t.Name()]; ok {
			if isPtr {
				*lines = append(*lines, ctx.initPtrMember("out"+fieldName, t))
				line = fmt.Sprintf("*(out%s) = DeepCopy%s(*(in%s))", fieldName, t.Name(), fieldName)
			} else {
				line = fmt.Sprintf("out%s = DeepCopy%s(in%s)", fieldName, t.Name(), fieldName)
			}
		} else {
			if isPtr {
				*lines = append(*lines, ctx.initPtrMember("out"+fieldName, t))
			}
			nl := ctx.genStructDeepCopyLines(t, lines, fieldName, iterLevel)
			needLength = needLength || nl
			line = "// end of " + fieldName
		}
	default:
		panic(fmt.Sprintf("Unknown Kind %s", t.Kind()))
	}
	*lines = append(*lines, line)
	return needLength
}


func (ctx *context) genStructDeepCopyLines(t reflect.Type, lines *[]string, fieldPrefix string, iterLevel int) bool {
	needLength := false
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		newPrefix := fieldPrefix+"."+field.Name
		nl := ctx.genFieldDeepCopyLines(field.Type, lines, newPrefix, iterLevel)
		needLength = needLength || nl
	}
	return needLength
}

