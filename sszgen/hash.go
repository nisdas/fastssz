package main

import (
	"fmt"
	"strings"
)

// hashTreeRoot creates a function that SSZ hashes the structs,
func (e *env) hashTreeRoot(name string, v *Value) string {
	tmpl := `// HashTreeRoot ssz hashes the {{.name}} object
	func (:: *{{.name}}) HashTreeRoot() ([]byte, error) {
		hh := ssz.DefaultHasherPool.Get()
		if err := ::.HashTreeRootWith(hh); err != nil {
			ssz.DefaultHasherPool.Put(hh)
			return nil, err
		}
		ssz.DefaultHasherPool.Put(hh)
		return nil, nil
	}
	
	// HashTreeRootWith ssz hashes the {{.name}} object with a hasher	
	func (:: *{{.name}}) HashTreeRootWith(hh *ssz.Hasher) error {
		{{.hashTreeRoot}}
		return nil
	}`

	data := map[string]interface{}{
		"name":         name,
		"hashTreeRoot": v.hashTreeRootContainer(true),
	}
	str := execTmpl(tmpl, data)
	return appendObjSignature(str, v)
}

func (v *Value) hashTreeRoot() string {
	switch v.t {
	case TypeContainer:
		return v.hashTreeRootContainer(false)

	case TypeBytes:
		if v.isFixed() && v.n == 32 {
			return fmt.Sprintf("hh.PutRoot(::.%s)", v.name)
		}
		return "// TODO BYTES"

	case TypeUint:
		return fmt.Sprintf("hh.PutUint64(::.%s)", v.name)

	case TypeBitList:
		return "// TODO BITLIST"

	case TypeBool:
		return fmt.Sprintf("hh.PutBool(::.%s)", v.name)

	case TypeVector:
		return "// TODO VECTOR"

	case TypeList:
		return "// TODO LIST"

	default:
		panic(fmt.Errorf("marshal not implemented for type %s", v.t.String()))
	}
}

func (v *Value) hashTreeRootContainer(start bool) string {
	if !start {
		return fmt.Sprintf("if err := ::.%s.HashTreeRootWith(hh); err != nil {\n return err\n}", v.name)
	}

	out := []string{}
	for indx, i := range v.o {
		str := fmt.Sprintf("// Field (%d) '%s'\n%s\n", indx, i.name, i.hashTreeRoot())
		out = append(out, str)
	}
	return strings.Join(out, "\n")
}
