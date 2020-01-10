package codon

import (
	"fmt"
	"reflect"
	"sort"
)

func getAllStructTypes(leafTypes map[string]string, t reflect.Type, name2type map[string]reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		switch field.Type.Kind() {
		case reflect.Struct:
			path := field.Type.PkgPath() + "." + field.Type.Name()
			if _, ok := leafTypes[path]; !ok {
				name2type[field.Type.Name()] = field.Type
				getAllStructTypes(leafTypes, field.Type, name2type)
			}
		case reflect.Ptr, reflect.Slice:
			t := field.Type
			if t.Elem().Kind() == reflect.Struct {
				path := t.Elem().PkgPath() + "." + t.Elem().Name()
				if _, ok := leafTypes[path]; !ok {
					name2type[t.Elem().Name()] = t.Elem()
					getAllStructTypes(leafTypes, t.Elem(), name2type)
				}
			}
		}
	}
}

func dumpProtoForMemberTypes(leafTypes map[string]string, indent string, t reflect.Type) {
}

func dumpField(leafTypes map[string]string, prefix string, fieldName string, fieldType reflect.Type, fieldNum int) {
	switch fieldType.Kind() {
	case reflect.Uintptr:
		panic("Uintptr is not supported")
	case reflect.Complex64:
		panic("complex64 is not supported")
	case reflect.Complex128:
		panic("complex128 is not supported")
	case reflect.Float32:
		panic("float32 is not supported")
	case reflect.Float64:
		panic("float64 is not supported")
	case reflect.Chan:
		panic("chan is not suported")
	case reflect.Func:
		panic("func is not suported")
	case reflect.Map:
		panic("map is not suported")
	case reflect.Bool:
		fmt.Printf(prefix+"bool %s = %d;\n", fieldName, fieldNum)
	case reflect.Int:
		fmt.Printf(prefix+"int64 %s = %d;\n", fieldName, fieldNum)
	case reflect.Int8:
		fmt.Printf(prefix+"int32 %s = %d;\n", fieldName, fieldNum)
	case reflect.Int16:
		fmt.Printf(prefix+"int32 %s = %d;\n", fieldName, fieldNum)
	case reflect.Int32:
		fmt.Printf(prefix+"int32 %s = %d;\n", fieldName, fieldNum)
	case reflect.Int64:
		fmt.Printf(prefix+"int64 %s = %d;\n", fieldName, fieldNum)
	case reflect.Uint:
		fmt.Printf(prefix+"uint64 %s = %d;\n", fieldName, fieldNum)
	case reflect.Uint8:
		fmt.Printf(prefix+"uint32 %s = %d;\n", fieldName, fieldNum)
	case reflect.Uint16:
		fmt.Printf(prefix+"uint32 %s = %d;\n", fieldName, fieldNum)
	case reflect.Uint32:
		fmt.Printf(prefix+"uint32 %s = %d;\n", fieldName, fieldNum)
	case reflect.Uint64:
		fmt.Printf(prefix+"uint64 %s = %d;\n", fieldName, fieldNum)
	case reflect.Struct:
		path := fieldType.PkgPath() + "." + fieldType.Name()
		if _, ok := leafTypes[path]; ok {
			fmt.Printf(prefix+"bytes %s = %d;\n", fieldName, fieldNum)
		} else if len(fieldType.Name()) == 0 {
			fmt.Printf(prefix+"%s %s = %d;\n", fieldName, fieldName, fieldNum)
		} else {
			fmt.Printf(prefix+"%s %s = %d;\n", fieldType.Name(), fieldName, fieldNum)
		}
	case reflect.Interface:
		fmt.Printf(prefix+"%s %s = %d;\n", fieldType.Name(), fieldName, fieldNum)
	case reflect.Ptr:
		path := fieldType.Elem().PkgPath() + "." + fieldType.Elem().Name()
		if _, ok := leafTypes[path]; ok {
			fmt.Printf(prefix+"bytes %s = %d;\n", fieldName, fieldNum)
		} else if len(fieldType.Name()) == 0 {
			fmt.Printf(prefix+"%s %s = %d;\n", fieldName, fieldName, fieldNum)
		} else {
			fmt.Printf(prefix+"%s %s = %d;\n", fieldType.Name(), fieldName, fieldNum)
		}
	case reflect.Array:
		if fieldType.Elem().Kind() == reflect.Uint8 {
			fmt.Printf(prefix+"bytes %s = %d;\n", fieldName, fieldNum)
		} else {
			panic("Only ByteArray is supported")
		}
	case reflect.Slice:
		panic("Should not reach here")
	case reflect.String:
		fmt.Printf(prefix+"string %s = %d;\n", fieldName, fieldNum)
	default:
		panic("not suported")
	}
}

func dumpProto(leafTypes map[string]string, indent string, t reflect.Type) {
	if t.Kind() != reflect.Struct {
		panic("Only accept struct types")
	}
	if structHasPrivateField(t) {
		panic("Cannot support structs with private fields")
	}
	fmt.Printf(indent+"message %s {\n", t.Name())
	dumpProtoForMemberTypes(leafTypes, indent+"    ", t)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldNum := i + 1
		if field.Type.Kind() == reflect.Slice {
			if field.Type.Elem().Kind() == reflect.Uint8 {
				fmt.Printf(indent+"    bytes %s = %d;\n", field.Name, fieldNum)
			} else {
				t := field.Type.Elem()
				if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
					fmt.Printf(indent+"    repeated bytes %s = %d;\n", field.Name, fieldNum)
				} else {
					prefix := indent+"    repeated "
					dumpField(leafTypes, prefix, field.Name, t, fieldNum)
				}
			}
		} else {
			dumpField(leafTypes, indent+"    ", field.Name, field.Type, fieldNum)
		}
	}
	fmt.Printf(indent+"} // %s\n\n", t.Name())
}

func (ctx *context) dumpIfcProto() {
	ifcPathList := make([]string, 0, len(ctx.ifcPath2Type))
	for name := range ctx.ifcPath2Type {
		ifcPathList = append(ifcPathList, name)
	}
	sort.Strings(ifcPathList)

	for _, ifcPath := range ifcPathList {
		alias := ctx.ifcPath2Alias[ifcPath]
		fmt.Printf("message %s {\n", alias)
		fmt.Printf("    oneof %s {\n", alias+"_impl")
		for _, structPath := range ctx.ifcPath2StructPaths[ifcPath] {
			alias := ctx.structPath2Alias[structPath]
			magicNum := ctx.structAlias2MagicNum[alias]
			fmt.Printf("        %s %s = %d;\n", alias, alias+"_var", magicNum)
		}
		fmt.Printf("    }\n")
		fmt.Printf("}\n")
	}
}

func (ctx *context) dumpStructProto(typeEntryList []TypeEntry) {
	name2type := make(map[string]reflect.Type)
	for _, entry := range typeEntryList {
		t := derefPtr(entry.Value)
		path := t.PkgPath() + "." + t.Name()
		if _, ok := ctx.leafTypes[path]; ok {
			continue
		}
		name2type[t.Name()] = t
		if t.Kind() != reflect.Struct {
			continue
		}
		//fmt.Printf("// %s %s %s\n", entry.Alias, entry.Name, t.Name())
		getAllStructTypes(ctx.leafTypes, t, name2type)
	}

	names := make([]string, 0, len(name2type))
	for name := range name2type {
		//fmt.Printf("//// %s \n", name)
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t := name2type[name]
		if t.Kind() == reflect.Struct {
			dumpProto(ctx.leafTypes, "", t)
		} else if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
			fmt.Printf("message %s {\n", name)
			fmt.Printf("    bytes %s = %d;\n", name+"_var", 1)
			fmt.Printf("}\n")
		} else if t.Kind() == reflect.Slice {
			fmt.Printf("// %s is ignored (slice of %v)\n", name, t.Elem())
		} else if t.Kind() != reflect.Interface {
			fmt.Printf("message %s {\n", name)
			dumpField(ctx.leafTypes, "    ", name+"_var", t, 1)
			fmt.Printf("}\n")
		}
	}
}

func DumpProtoFile(
	// contains the types which should be regarded as leaf types
	// Key is the full type name, Value is the short type name
	leafTypes map[string]string,
	// Some struct->interface implementation relationship must be ignored
	// Key is struct's alias and Value is interface's alias
	ignoreImpl map[string]string,
	// The types for which we will generate code
	typeEntryList []TypeEntry) {

	// Now initialize the context
	ctx := newContext(leafTypes, ignoreImpl)
	for _, entry := range typeEntryList {
		ctx.register(entry.Alias, entry.Name, entry.Value)
	}
	ctx.analyzeIfc()

	fmt.Printf("syntax = \"proto3\";\n")
	ctx.dumpStructProto(typeEntryList)
	ctx.dumpIfcProto()
}

