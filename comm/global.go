package comm

var (
	//default base tables
	GlobalBaseTables = [...]string{"BaseVersion", "BaseGUID", "BaseSelfIncrementIDData", "BaseAccounts", "BaseRoles"}
	//default base table primary keys
	GlobalBaseTableMap = map[string]string{
		"BaseVersion":             "version",
		"BaseGUID":                "guid",
		"BaseSelfIncrementIDData": "id",
		"BaseAccounts":            "token",
		"BaseRoles":               "role_id",
	}
	//table base, pub and split proto files
	GlobalTableFiles = map[string]string{
		"BASE": "base.proto",
		"PUB":  "table_pub_message.proto",
		"OUT":  "table_split_message.proto",
		"IN":   "table_split_message.proto",
	}
	//blob proto files
	GlobalBlobFiles = map[string]string{
		"IN":  "blob_user_data_in.proto",
		"OUT": "blob_user_data_out.proto",
	}
	//import paths for ignoring, not parse
	GlobalIgnoreImportPaths = []string{
		"proto/entity/common.proto",
		"proto/entity/enumm_entity.proto",
	}
)
var (
	//business variables
	//base tables, read item `base_tables` from config file , if not exist in config file, assigned by default `GlobalBaseTables`
	BaseTables []string
	//base table primary keys map, read item `base_table_primary_keys` from config file, if not exist in config file, assigned by default `GlobalBaseTableMap`
	BaseTableMap map[string]string
	//base, pub and split proto files map, read item `pub_split_proto_files` from config file, if not exist in config file, assigned by default `GlobalPubSplitProtoFiles`
	TableFiles map[string]string
	//blob proto files map, read item `blob_proto_files` from config file, if not exist in config file, assigned by default `GlobalBlobFiles`
	BlobFiles map[string]string
	//import paths for ignoring, read item `import_path_ignores` from config file, if not exist in config file, assigned by default `GlobalIgnoreImportPaths`
	IgnoreImportPaths []string
)

var (
	//tcaplusdb constants
	//tcaplusdb entity package name, read item `tcaplus_package_name` from config file, if not exist in config, assigned by default
	TcaplusPackageName string = "tcaplus_entity"
	//tcaplusdb import path, read item `tcaplus_import_path` from config file, if not exist in config, assigned by default
	TcaplusImportName string = "tcaplusservice.optionv1.proto"
)

var (
	//business constants
	ProtoSuffix string = ".proto"
	//business entity package name
	CustomPackageName string = "entity"
	//blob message name, read item `blob_user_in_msg_name` from config file, if not exist in config, assigned by default
	BlobUserInMsg string = "blob_user_data_in"
	//blob message name, read item `blob_user_out_msg_name` from config file, if not exist in config, assigned by default
	BlobUserOutMsg string = "blob_user_data_out"
	//specifiy proto files for ignoring parsing, read item `proto_file_ignores` from config file, if not exist in config, assigned by default
	IgnoreProtoFiles string = ""

	CommonProtoFile string = "common.proto"
	EnumProtoFile   string = "enumm_entity.proto"
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
