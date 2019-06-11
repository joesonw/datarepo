package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	astgen "github.com/joesonw/go-ast-gen"
	"github.com/onsi/gomega"
)

func TestParseAction(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	die := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	fset := token.FileSet{}
	node, err := parser.ParseFile(&fset, "", `
	package main 
	type A struct {}
	type Test interface {
		FindEqualID(ctx context.Context, id int64) (*A, error)
		FindEqualType(ctx context.Context, typ string) ([]*A, error)
		FindOneEqualType(ctx context.Context, typ string) (*A, error)
		FindManyEqualType(ctx context.Context, typ string) ([]*A, error)
	}
	`, parser.ParseComments)
	die(err)

	methods, err := astgen.ParseInterface(node.Decls[1].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.InterfaceType), astgen.ImportAll)
	die(err)

	query, err := Parse(methods[0])
	die(err)
	g.Expect(query.Action).To(gomega.Equal(FindOne))

	query, err = Parse(methods[1])
	die(err)
	g.Expect(query.Action).To(gomega.Equal(FindMany))

	query, err = Parse(methods[2])
	die(err)
	g.Expect(query.Action).To(gomega.Equal(FindOne))

	query, err = Parse(methods[3])
	die(err)
	g.Expect(query.Action).To(gomega.Equal(FindMany))
}
