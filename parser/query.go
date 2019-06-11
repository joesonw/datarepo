package parser

import (
	"fmt"
	"strings"
)

type Query struct {
	Action   Action
	Negate   bool
	Group    Group
	Sorts    []Sort
	PageName string
}

func (q Query) String() string {
	sorts := make([]string, len(q.Sorts))
	for i := range q.Sorts {
		sorts[i] = q.Sorts[i].String()
	}
	not := ""
	if q.Negate {
		not = "NOT "
	}
	return fmt.Sprintf("%s %s%s %s", q.Action, not, q.Group.String(), strings.Join(sorts, ","))
}

type Sort struct {
	Key   string
	Order Order
}

func (s Sort) String() string {
	return fmt.Sprintf("ORDER BY %s %s", s.Key, s.Order.String())
}

type Group struct {
	Logic Logic
	Pairs []Pair
}

func (g Group) String() string {
	pairs := make([]string, len(g.Pairs))
	for i := range g.Pairs {
		pairs[i] = g.Pairs[i].String()
	}
	return strings.Join(pairs, " "+g.Logic.String()+" ")
}

type Pair struct {
	Key      string
	Operator Operator
	Name     string
}

func (p Pair) String() string {
	return fmt.Sprintf("\"%s\" %s ?", p.Key, p.Operator)
}
