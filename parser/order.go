package parser

type Order string

var (
	Ascend  Order = "ASC"
	Descend Order = "DESC"

	AllOrders = map[Order]bool{
		Ascend:  true,
		Descend: true,
	}

	OrderNames = map[Order][]string{
		Ascend:  {"Ascend", "Asc", "ASC"},
		Descend: {"Descend", "Desc", "DESC"},
	}
)

func (o Order) String() string {
	return string(o)
}

func (o Order) Valid() bool {
	return AllOrders[o]
}
