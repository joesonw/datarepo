package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"text/template"
	"time"

	"github.com/joesonw/datarepo/generator"

	repoParser "github.com/joesonw/datarepo/parser"
	astgen "github.com/joesonw/go-ast-gen"
	"github.com/spf13/pflag"
)

var (
	pEngine = pflag.StringP("engine", "e", "mongo", "engine type")
	pPage   = pflag.StringP("page", "p", "", "page request name")

	templateRoot = template.Must(template.New("").Parse(`
		import (
			"context"

			"go.mongodb.org/mongo-driver/mongo/options"
			"go.mongodb.org/mongo-driver/bson/primitive"
			"go.mongodb.org/mongo-driver/bson"
			"go.mongodb.org/mongo-driver/mongo"
		)

		type {{.Name}}Repository struct {
			collection *mongo.Collection
		}

		func New{{.Name}}Repository(collection *mongo.Collection) *{{.Name}}Repository {
			return &{{.Name}}Repository{
				collection,
			}
		}
	`))
)

func main() {
	pflag.Parse()
	engine := *pEngine
	page := *pPage

	if engine != "mongo" {
		die(fmt.Errorf("-engine \"%s\" is not supported", engine))
	}

	if page == "" {
		die(errors.New("-page is requried"))
	}

	goPackage := os.Getenv("GOPACKAGE")
	goFile := os.Getenv("GOFILE")

	cwd, err := os.Getwd()
	die(err)
	file := path.Join(cwd, goFile)

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	die(err)

	goLine, err := strconv.ParseInt(os.Getenv("GOLINE"), 10, 64)
	die(err)

	var spec *ast.TypeSpec
	for _, d := range node.Decls {
		decl, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}

		spec, ok = decl.Specs[0].(*ast.TypeSpec)
		if !ok {
			continue
		}

		if fset.Position(d.Pos()).Line < int(goLine) {
			continue
		}
		break
	}

	if spec == nil {
		panic("not found")
	}
	typeName := spec.Name.Name

	output := fmt.Sprintf("package %s\n\n", goPackage)
	output += fmt.Sprintf("// GENERATED CODE BY github.com/joesonw/repogen, PLEASE DO NOT MODIFY. at %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	interfaceType := spec.Type.(*ast.InterfaceType)
	methods, err := astgen.ParseInterface(interfaceType, astgen.ImportAll)
	die(err)

	buffer := bytes.NewBuffer(nil)
	die(templateRoot.Execute(buffer, map[string]interface{}{
		"Name": typeName,
	}))
	output += buffer.String()

	var gen generator.Interface
	gen = generator.Mongo{}

	for _, method := range methods {
		query, err := repoParser.Parse(method, page)
		die(err)
		result, err := gen.Generate(method, query)
		die(err)
		output += fmt.Sprintf("func (z *%sRepository) %s {\n", typeName, method.String())
		output += result + "\n"
		output += "}\n"
	}

	out := path.Join(cwd, goFile[:len(goFile)-3]+"_repository_gen.go")
	die(ioutil.WriteFile(out, []byte(output), 0644))
	die(exec.Command("go", "fmt", out).Run())
	formatted, err := exec.Command("goimports", out).Output()
	die(err)
	die(ioutil.WriteFile(out, formatted, 0644))
}

func die(err error) {
	if err != nil {
		panic(err)
	}
}
