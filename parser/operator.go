package parser

type Operator string

var (
	In          Operator = "IN"
	NotIn       Operator = "NotIN"
	Equal       Operator = "EQUAL"
	NotEqual    Operator = "NotEQUAL"
	MoreThan    Operator = "MoreTHAN"
	MoreOrEqual Operator = "MoreOrEQUAL"
	LessThan    Operator = "LessTHAN"
	LessOrEqual Operator = "LessOrEQUAL"
	Like        Operator = "LIKE"

	AllOperators = map[Operator]bool{
		In:          true,
		NotIn:       true,
		NotEqual:    true,
		Equal:       true,
		MoreThan:    true,
		MoreOrEqual: true,
		LessThan:    true,
		LessOrEqual: true,
		Like:        true,
	}

	OperatorNames = map[Operator][]string{
		In:          {"In", "Among"},
		NotIn:       {"NotIn", "Nin", "NotAmong"},
		NotEqual:    {"NotEqual"},
		Equal:       {"Equal", "Is", "Be", "By"},
		MoreThan:    {"MoreThan", "GreaterThan"},
		MoreOrEqual: {"MoreEqualThan", "MoreEqual", "GreaterEqualThan", "GreaterEqual"},
		LessThan:    {"LessThan"},
		LessOrEqual: {"LessEqualThan", "LessEqual"},
		Like:        {"Like"},
	}
)

func (o Operator) String() string {
	return string(o)
}

func (o Operator) Valid() bool {
	return AllOperators[o]
}
