package generator

import (
	"github.com/joesonw/datarepo/parser"
	astgen "github.com/joesonw/go-ast-gen"
)

type Interface interface {
	Generate(method astgen.Method, query parser.Query) (string, error)
}
