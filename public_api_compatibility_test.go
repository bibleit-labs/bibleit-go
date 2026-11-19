package bibleit

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The baseline records the public v0.4.0 source contract, including enum names.
// Schema generation may add declarations and optional struct fields, but must
// not silently rename existing constants when a new enum uses the same value.
func TestPublicAPICompatibility(t *testing.T) {
	var baseline map[string]string
	data, err := os.ReadFile("contract/compatibility/go-v0.4.0-api.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	current, err := publicAPISnapshot(".")
	if err != nil {
		t.Fatal(err)
	}
	for key, expected := range baseline {
		actual, ok := current[key]
		if !ok {
			t.Errorf("removed public declaration %s", key)
		} else if actual != expected {
			t.Errorf("changed public declaration %s: want %q, got %q", key, expected, actual)
		}
	}
}
func publicAPISnapshot(directory string) (map[string]string, error) {
	set := token.NewFileSet()
	out := map[string]string{}
	files, err := filepath.Glob(filepath.Join(directory, "*.go"))
	if err != nil {
		return nil, err
	}
	render := func(value any) string { var b bytes.Buffer; _ = format.Node(&b, set, value); return b.String() }
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			return nil, err
		}
		for _, declaration := range file.Decls {
			switch d := declaration.(type) {
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				name := d.Name.Name
				if d.Recv != nil {
					name = "(" + render(d.Recv.List[0].Type) + ")." + name
				}
				out["func:"+name] = render(d.Type)
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if !s.Name.IsExported() {
							continue
						}
						kind := "type:"
						if s.Assign.IsValid() {
							kind = "alias:"
						}
						if body, ok := s.Type.(*ast.StructType); ok {
							out[kind+s.Name.Name] = "struct"
							for _, field := range body.Fields.List {
								for _, name := range field.Names {
									if name.IsExported() {
										tag := ""
										if field.Tag != nil {
											tag = " " + field.Tag.Value
										}
										out["field:"+s.Name.Name+"."+name.Name] = render(field.Type) + tag
									}
								}
							}
						} else {
							out[kind+s.Name.Name] = render(s.Type)
						}
					case *ast.ValueSpec:
						for i, name := range s.Names {
							if name.IsExported() {
								value := render(s.Type)
								if len(s.Values) > i {
									value += "=" + render(s.Values[i])
								}
								out[d.Tok.String()+":"+name.Name] = value
							}
						}
					}
				}
			}
		}
	}
	return out, nil
}
