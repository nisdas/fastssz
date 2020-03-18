package main

import (
	"fmt"
	"strings"
)

// hashTreeRoot creates a function that SSZ hashes the structs,
func (e *env) hashTreeRoot(name string, v *Value) string {
	tmpl := `// HashTreeRoot ssz hashes the {{.name}} object
	func (:: *{{.name}}) HashTreeRoot() ([]byte, error) {
		return ssz.HashWithDefaultHasher(::)
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
		if v.isFixed() {
			if v.n == 32 {
				return fmt.Sprintf("hh.PutRoot(::.%s)", v.name)
			} else if v.n < 32 {
				return fmt.Sprintf("hh.PutFixedBytes(::.%s)", v.name)
			}
		}
		return fmt.Sprintf("if err := hh.PutBytes(::.%s); err != nil {\nreturn err\n}", v.name)

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

	tmpl := `hh.Bound()

	{{.fields}}

	if err := hh.BitwiseMerkleize(); err != nil {
		return err
	}`

	return execTmpl(tmpl, map[string]interface{}{
		"fields": strings.Join(out, "\n"),
	})
}
