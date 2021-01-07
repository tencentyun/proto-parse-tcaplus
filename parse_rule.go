package main

type Import struct {
	Path string
}

type Package struct {
	Name string
}

type Syntax struct {
	Name string
}

type Option struct {
	Name       string
	Value      string
	Aggregated []Option
}

type Message struct {
	Name          string
	Fields        []Field
	Maps          []Map
	ReservedIDs   []int
	ReservedNames []string
	Messages      []Message
	Options       []Option
	Enums         []Enum
}

type EnumField struct {
	Name    string
	Integer int
	Options []Option
}

type Enum struct {
	Name          string
	EnumFields    []EnumField
	ReservedIDs   []int
	ReservedNames []string
	AllowAlias    bool
}

type Map struct {
	KeyType string `json:"key_type,omitempty"`
	Field   Field
}

type Field struct {
	ID         int
	Name       string
	Type       string
	IsRepeated bool
	Options    []Option
}
