# proto-parse-tcaplus

convert business proto files to tcaplusdb proto files

# Compile

- Execute `make` command to compile the tool
- All compiled results are in `bin` directory

# Usage

```
Usage:
  proto-parse-tcaplus [flags]

Examples:
  ./proto-parse-tcaplus -s "./testdata/test" -d "./out/test"  -c "./config/proto_parse.cfg

Flags:
  -c, --config string        tool config file
  -d, --dest-path string     destination path of generated proto files
  -h, --help                 help for proto-parse-tcaplus
  -s, --source-path string   source path of proto files
```

**Parameters**:

- **-s**: source proto files that need to be converted, refer to `testdata/test` directory.

- **-d**: dest proto files that are converted from source proto files, all proto files will be converted into five proto files, such as `base.proto, blob_user_data_in.proto, blob_user_data_out.proto, table_pub_message.proto, table_split_message.proto`
- **-c**: config file that contains business configs and common configs

# Config

Demo config file is as below:

```
[business]
    #business base table, comma separate
    base_tables = "BaseVersion, BaseGUID, BaseSelfIncrementIDData, BaseAccounts, BaseRoles"
    #base table primary keys, comma separate each table, ':' separates table and primary key, if table has multiple primary keys, use # to separate
    base_table_primary_keys = "BaseVersion:version, BaseGUID:guid:uid, BaseSelfIncrementIDData:id, BaseAccounts:token, BaseRoles:roleID"
    #pub, split proto
    table_proto_files = "BASE:base.proto, PUB:table_pub_message.proto, SPLIT:table_split_message.proto"
    #blob proto files, comma separate each proto file, `:` separates blob type and proto file name
    blob_proto_files = "IN:blob_user_data_in.proto, OUT:blob_user_data_out.proto"
    #blob user in msg name
    blob_user_in_msg_name = "BlobUserDataIn"
    # blob user out msg name
    blob_user_out_msg_name = "BlobUserDataOut"
    #ignore parse specified proto files, comma separate eacch proto file
    proto_file_ignores = ""
    #ignore import paths, comma separates each import path
    import_path_ignores = "proto/entity/common.proto, proto/entity/enumm_entity.proto"

[tcaplusdb]
    # tcaplusdb entity package name
    tcaplus_package_name = "tcaplus_entity"
    # tcaplusdb import path name
    tcaplus_import_path = "tcaplusservice.optionv1.proto"
```

- **base_tables**: Setup the basic tables of business, six tables by default.
- **base_table_primary_keys**: Specify the primary keys of each basic table, support specifying multiple primary keys for each table, and using comma to separate them.
- **table_proto_files**: Specify the output proto files for `BASE`, `PUB`, and `SPLIT` messages.
- **blob_proto_files**: Specify the output proto files for `BLOB` messages.
- **blob_user_in_msg_name**: Specify the proto file for IN blob messages.
- **blob_user_out_msg_name**: Specify the proto file for OUT blob messages.
- **proto_file_ignores**: Specify the proto files that ignores parsing.
- **import_path_ignores**: Specify the import path that ignores importing.
- **tcaplus_package_name**: Specify the package name of tcaplusdb interfaces
- **tcaplus_import_path**: The dedicated import path of tcaplusdb proto file.
