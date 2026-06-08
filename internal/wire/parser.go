package wire

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type Dependency struct {
	FieldName string
	TypeName  string // The struct type it depends on (e.g. "*UserService" -> "UserService")
}

type StructNode struct {
	Name         string
	Dependencies []Dependency
}

// ParseDir scans a directory for structs and their inject dependencies.
func ParseDir(dirPath string) ([]StructNode, string, error) {
	fset := token.NewFileSet()

	// Parse all .go files in the directory
	pkgs, err := parser.ParseDir(fset, dirPath, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go") && !strings.HasSuffix(info.Name(), "gox_wire_gen.go")
	}, parser.ParseComments)

	if err != nil {
		return nil, "", err
	}

	var nodes []StructNode
	var pkgName string

	for name, pkg := range pkgs {
		pkgName = name // Assume one package per directory for simplicity
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				ts, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}

				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					return true
				}

				node := StructNode{Name: ts.Name.Name}
				if st.Fields != nil {
					for _, field := range st.Fields.List {
						if field.Tag != nil && strings.Contains(field.Tag.Value, `inject:""`) {
							// Find the type name
							var typeName string
							switch t := field.Type.(type) {
							case *ast.StarExpr:
								if ident, ok := t.X.(*ast.Ident); ok {
									typeName = ident.Name
								} else if sel, ok := t.X.(*ast.SelectorExpr); ok {
									typeName = sel.X.(*ast.Ident).Name + "." + sel.Sel.Name
								}
							case *ast.Ident:
								typeName = t.Name
							case *ast.SelectorExpr:
								typeName = t.X.(*ast.Ident).Name + "." + t.Sel.Name
							}

							if typeName != "" {
								for _, name := range field.Names {
									node.Dependencies = append(node.Dependencies, Dependency{
										FieldName: name.Name,
										TypeName:  typeName,
									})
								}
							}
						}
					}
				}

				nodes = append(nodes, node)
				
				return false
			})
		}
	}

	return nodes, pkgName, nil
}
