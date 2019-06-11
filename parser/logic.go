package parser

type Logic string

var (
	And Logic = "AND"
	Or  Logic = "OR"

	AllLogics = map[Logic]bool{
		And: true,
		Or:  true,
	}
)

func (l Logic) String() string {
	return string(l)
}

func (l Logic) Valid() bool {
	return AllLogics[l]
}
