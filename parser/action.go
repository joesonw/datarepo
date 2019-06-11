package parser

type Action string

var (
	FindOne    Action = "FindOne"
	FindMany   Action = "FindMany"
	Replace    Action = "Replace"
	Create     Action = "Create"
	RemoveOne  Action = "RemoveOne"
	RemoveMany Action = "RemoveMany"
	Count      Action = "Count"

	AllActions = map[Action]bool{
		FindOne:    true,
		FindMany:   true,
		Replace:    true,
		Create:     true,
		RemoveMany: true,
		RemoveOne:  true,
		Count:      true,
	}

	ActionNames = map[Action][]string{
		Count:      {"Count", "Len"},
		FindOne:    {"FindOne", "One"},
		FindMany:   {"FindMany", "Many"},
		Replace:    {"Replace", "Save"},
		Create:     {"Create"},
		RemoveOne:  {"RemoveOne"},
		RemoveMany: {"RemoveMany"},
	}
)

func (a Action) String() string {
	return string(a)
}

func (a Action) Valid() bool {
	return AllActions[a]
}
