package main

var (
	//business variables
	FixTables   = [...]string{"BaseVersion", "BaseGUID", "BaseSelfIncrementIDData", "BaseAccounts", "BaseRoles"}
	FixTableMap = map[string]string{
		"BaseVersion":             "version",
		"BaseGUID":                "guid",
		"BaseSelfIncrementIDData": "id",
		"BaseAccounts":            "token",
		"BaseRoles":               "role_id",
	}
)

const (
	//tcaplusdb constants
	TcaplusPackageName string = "tcaplus_entity"
	TcaplusImportName  string = "tcaplusservice.optionv1.proto"
)

const (
	//business constants
	ProtoSuffix       string = ".proto"
	CustomPackageName string = "entity"

	BlobUserInMsg    string = "blob_user_data_in"
	BlobUserOutMsg   string = "blob_user_data_out"
	IgnoreProtoFiles string = "common.proto,enumm_entity.proto"
	CommonProtoFile  string = "common.proto"
	EnumProtoFile    string = "enumm_entity.proto"
)

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
