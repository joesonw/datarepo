package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	astgen "github.com/joesonw/go-ast-gen"

	"github.com/joesonw/datarepo/parser"
)

var (
	templateCreate = template.Must(template.New("").Parse(`
		result, err := z.collection.InsertOne({{.Context}}, {{.In}})
		if err != nil {
			return err
		}
		{{.In}}.ID = result.InsertedID.(primitive.ObjectID)
		return nil
	`))
	templateRemoveOne = template.Must(template.New("").Parse(`
		_, err = z.collection.DeleteOne({{.Context}}, {{.Query}})
		return err
	`))
	templateRemoveMany = template.Must(template.New("").Parse(`
		_, err = z.collection.DeleteMany({{.Context}}, {{.Query}})
		return err
	`))
	templateReplace = template.Must(template.New("").Parse(`
		_, err = z.collection.ReplaceOne({{.Context}}, {{.Query}}, {{.In}})
		return err
	`))
	templateFindOne = template.Must(template.New("").Parse(`
		{{.Variables}}
		result := &{{.Type}}{}
		err := z.collection.FindOne({{.Context}}, {{.Query}}, {{.Options}}).Decode(result)
		if err != nil {
			return nil, err
		}
		return result, nil 
	`))
	templateFindMany = template.Must(template.New("").Parse(`
		{{.Variables}}
		cursor, err := z.collection.Find({{.Context}}, {{.Query}}, {{.Options}})
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)
		var list []*{{.Type}}
		for cursor.Next(ctx) {
			in := &{{.Type}}{}
			if err := cursor.Decode(in); err != nil {
				return nil, err
			}
			list = append(list, in)
		}
		return list, err
	`))
	templateCount = template.Must(template.New("").Parse(`
		return z.collection.CountDocuments({{.Context}}, {{.Query}})
	`))
)

type m map[string]interface{}

type Mongo struct {
}

func (mongo Mongo) Generate(method astgen.Method, query parser.Query) (string, error) {
	switch query.Action {
	case parser.Create:
		return mongo.runTemplate(templateCreate, m{
			"Context": method.Ins()[0].Name(),
			"In":      method.Ins()[1].Name(),
		})
	case parser.Replace:
		return mongo.runTemplate(templateReplace, m{
			"Context": method.Ins()[0].Name(),
			"In":      method.Ins()[1].Name(),
			"Query":   mongo.generateQuery(query),
		})
	case parser.RemoveOne:
		return mongo.runTemplate(templateRemoveOne, m{
			"Context": method.Ins()[0].Name(),
			"Query":   mongo.generateQuery(query),
		})
	case parser.RemoveMany:
		return mongo.runTemplate(templateRemoveMany, m{
			"Context": method.Ins()[0].Name(),
			"Query":   mongo.generateQuery(query),
		})
	case parser.Count:
		return mongo.runTemplate(templateCount, m{
			"Context": method.Ins()[0].Name(),
			"Query":   mongo.generateQuery(query),
		})
	case parser.FindOne:
		{
			options := "nil"
			variables := ""
			if query.PageName != "" {
				variables = fmt.Sprintf("pageSkip := %s.Page * %s.Size", query.PageName, query.PageName)
				options = fmt.Sprintf("&options.FindOneOptions{Limit:&%s.Size,Skip: &pageSkip}", query.PageName)
			}
			return mongo.runTemplate(templateFindOne, m{
				"Context":   method.Ins()[0].Name(),
				"Query":     mongo.generateQuery(query),
				"Type":      method.Outs()[0].Type().Types()[0].String(),
				"Options":   options,
				"Variables": variables,
			})
		}
	case parser.FindMany:
		{
			options := "nil"
			variables := ""
			if query.PageName != "" {
				variables = fmt.Sprintf("pageSkip := %s.Page * %s.Size", query.PageName, query.PageName)
				options = fmt.Sprintf("&options.FindOptions{Limit:&%s.Size,Skip: &pageSkip}", query.PageName)
			}
			return mongo.runTemplate(templateFindMany, m{
				"Context":   method.Ins()[0].Name(),
				"Query":     mongo.generateQuery(query),
				"Type":      method.Outs()[0].Type().Types()[0].Types()[0].String(),
				"Options":   options,
				"Variables": variables,
			})
		}
	}
	return "", nil
}

func (Mongo) generateQuery(query parser.Query) string {
	paris := make([]string, len(query.Group.Pairs))
	for i, pair := range query.Group.Pairs {
		if pair.Operator == parser.Equal {
			paris[i] = fmt.Sprintf("bson.M{\n\"%s\": %s,\n}", pair.Key, pair.Name)
			continue
		}
		operator := ""
		value := pair.Name
		switch pair.Operator {
		case parser.In:
			operator = "$in"
		case parser.NotIn:
			operator = "$nin"
		case parser.NotEqual:
			operator = "$ne"
		case parser.MoreThan:
			operator = "$gt"
		case parser.MoreOrEqual:
			operator = "$gte"
		case parser.LessThan:
			operator = "$lt"
		case parser.LessOrEqual:
			operator = "$lte"
		case parser.Like:
			operator = "$regex"
			value = fmt.Sprintf("\"/\"+%s+\"/\"", value)
		}
		paris[i] = fmt.Sprintf("bson.M{\n\"%s\": bson.M{\"%s\": %s},\n}", pair.Key, operator, value)
	}

	logic := "$and"
	if query.Group.Logic == parser.Or {
		logic = "$or"
	}

	stmt := strings.Join(paris, ",\n")
	if len(query.Group.Pairs) > 1 {
		stmt = fmt.Sprintf("bson.M{\n\"%s\":bson.A{\n%s,\n},\n}", logic, stmt)
	}

	if query.Negate {
		stmt = fmt.Sprintf("bson.M{\n\"$not\":%s\n}", stmt)
	}
	if len(query.Group.Pairs) == 0 {
		stmt = ""
	}

	if len(query.Sorts) > 0 {
		orderBy := "bson.M{"
		for _, sort := range query.Sorts {
			order := "1"
			if sort.Order == parser.Descend {
				order = "-1"
			}
			orderBy += fmt.Sprintf("\"%s\": %s", sort.Key, order)
		}
		orderBy += "}"
		if stmt != "" {
			stmt = fmt.Sprintf("\n\"$query\":%s,", stmt)
		}
		stmt = fmt.Sprintf("bson.M{%s\n\"$orderBy\":%s,\n}", stmt, orderBy)
	}
	if stmt == "" {
		stmt = "nil"
	}
	return stmt
}

func (Mongo) runTemplate(t *template.Template, data map[string]interface{}) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := t.Execute(buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
