package parser

import (
	"errors"
	"fmt"
	"strings"

	astgen "github.com/joesonw/go-ast-gen"
)

func Parse(method astgen.Method, pageRequestName string) (query Query, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %s", method.String(), err.Error())
		}
	}()

	name := method.Name()
	if strings.HasPrefix(name, "Exec") {
		return
	}

	resultType := method.Outs()[0].Type()
	numOfParameters := len(method.Ins()) - 1
	beginIndex := 1

	for action, actionNames := range ActionNames {
		found := false
		for _, actionName := range actionNames {
			if strings.HasPrefix(name, actionName) {
				name = name[len(actionName):]
				query.Action = action
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !query.Action.Valid() && strings.HasPrefix(name, "Find") {
		if resultType.Kind() == astgen.Slice {
			query.Action = FindMany
		} else {
			query.Action = FindOne
		}
		name = name[4:]
	}

	if query.Action == Count && resultType.String() != "int64" {
		err = errors.New("Count should have int64 as result")
		return
	}

	if !query.Action.Valid() {
		err = fmt.Errorf("no valid action")
		return
	}

	/************ BEGIN Check Return Values ****************/

	if query.Action == FindOne && resultType.Kind() == astgen.Slice {
		err = fmt.Errorf("with Action \"%s\", result value must not be slice", query.Action)
		return
	}

	if query.Action == FindMany && resultType.Kind() != astgen.Slice {
		err = fmt.Errorf("with Action \"%s\", result value must be slice", query.Action)
		return
	}

	if query.Action == Create || query.Action == Replace {
		beginIndex = 2
		if len(method.Outs()) != 1 || method.Outs()[0].Type().String() != "error" {
			err = fmt.Errorf("must have exeact one return value and is error")
			return
		}
	} else {
		if len(method.Outs()) != 2 || method.Outs()[1].Type().String() != "error" {
			err = fmt.Errorf("must have two return values, and second being error")
			return
		}
	}
	/************ END Check Return Values ****************/

	if strings.Index(name, "OrderBy") >= 0 {
		segs := strings.Split(name, "OrderBy")
		if len(segs) != 2 {
			err = fmt.Errorf("wrong numebr of OrderBy appereance")
			return
		}
		name = segs[0]
		keys := strings.Split(segs[1], "And")
		for _, key := range keys {
			sort := Sort{}
			for order, orderNames := range OrderNames {
				found := false
				for _, orderName := range orderNames {
					if strings.HasSuffix(key, orderName) {
						sort.Order = order
						sort.Key = key[:len(key)-len(orderName)]
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if sort.Order == "" {
				sort.Order = Ascend
				sort.Key = key
			}
			if !sort.Order.Valid() {
				err = fmt.Errorf("invalid order")
				return
			}
			query.Sorts = append(query.Sorts, sort)
		}
	}

	if query.Action == Replace || query.Action == Create || query.Action == RemoveOne || query.Action == RemoveMany {
		if len(query.Sorts) != 0 {
			err = fmt.Errorf("should not have sort")
			return
		}
	}

	if query.Action == Create {
		numOfParameters = 0
	} else if query.Action == Replace || query.Action == RemoveOne || query.Action == RemoveMany {
		numOfParameters -= 1
	}

	name = strings.ReplaceAll(name, "MoreOrEqual", "MoreEqual")
	name = strings.ReplaceAll(name, "LessOrEqual", "LessEqual")
	name = strings.ReplaceAll(name, "GreaterOrEqual", "GreaterEqual")

	if strings.Index(name, "And") >= 0 && strings.Index(name, "Or") >= 0 {
		err = fmt.Errorf("query can only be concated or \"AND\" or \"OR\", not both")
		return
	}

	if strings.HasPrefix(name, "Not") && !strings.HasPrefix(name, "NotEqual") {
		query.Negate = true
		name = name[3:]
	}

	if len(name) > 0 {
		var keys []string

		if strings.Index(name, "And") >= 0 {
			query.Group.Logic = And
			keys = strings.Split(name, "And")
		} else if strings.Index(name, "Or") >= 0 {
			query.Group.Logic = Or
			keys = strings.Split(name, "Or")
		} else {
			keys = []string{name}
		}

		for i, key := range keys {
			pair := Pair{}
			pair.Name = method.Ins()[beginIndex+i].Name()
			for operator, operatorNames := range OperatorNames {
				found := false
				for _, operatorName := range operatorNames {
					if strings.HasPrefix(key, operatorName) {
						key = key[len(operatorName):]
						pair.Operator = operator
						found = true
						break
					}
				}
				if found {
					break
				}
			}

			isSlice := method.Ins()[i+1].Type().Kind() == astgen.Slice
			if pair.Operator == NotIn || pair.Operator == In {
				if !isSlice {
					err = fmt.Errorf("param %s must be of slice", key)
					return
				}
			} else {
				if isSlice {
					err = fmt.Errorf("param %s must not be of slice", key)
					return
				}
			}

			if !pair.Operator.Valid() {
				err = fmt.Errorf("invalid operator")
				return
			}

			if strings.Index(key, "Of") >= 0 {
				parts := strings.Split(key, "Of")
				key = parts[len(parts)-1]
				for i := len(parts) - 2; i >= 0; i-- {
					key += "." + parts[i]
				}
			}

			pair.Key = key
			query.Group.Pairs = append(query.Group.Pairs, pair)
		}
	}

	if query.Action == Create && len(query.Group.Pairs) != 0 {
		err = errors.New("Create should not have query")
		return
	}

	lastParam := method.Ins()[len(method.Ins())-1]
	if lastParam.Type().String() == pageRequestName {
		numOfParameters -= 1
		query.PageName = lastParam.Name()
	}

	if query.Action != FindOne && query.Action != FindMany && query.PageName != "" {
		err = fmt.Errorf("Action \"%s\" should not have page", query.Action)
		return
	}

	if len(query.Group.Pairs) != numOfParameters {
		err = fmt.Errorf("number of parameters does not fulfill method")
		return
	}

	return
}
