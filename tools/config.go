package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/tencentyun/proto-parse-tcaplus/comm"
	"gopkg.in/ini.v1"
)

func ReadIni(iniFile string) (*ini.File, error) {

	if _, err := os.Stat(iniFile); os.IsNotExist(err) {
		fmt.Println(iniFile + " is no exist")
		return nil, err
	}

	cfg, err := ini.Load(iniFile)
	//	fmt.Println("ini load error :" + err.Error())
	return cfg, err

}

func ParseCfg(cfg *ini.File) error {
	if cfg == nil {
		fmt.Println("cfg is nil")
	}
	busSec, err := cfg.GetSection("business")
	if err != nil {
		fmt.Println(err)
		return err
	}

	//init map object
	comm.FixTableMap = make(map[string]string)
	comm.BlobFiles = make(map[string]string)

	if ok := busSec.HasKey("base_tables"); ok {
		//parse base tables
		baseTables := strings.Split(busSec.Key("base_tables").Value(), ",")
		for i := range baseTables {
			baseTables[i] = strings.TrimSpace(baseTables[i])
			comm.FixTables = append(comm.FixTables, baseTables[i])
		}

	} else {
		//init default values
		comm.FixTables = append(comm.FixTables, comm.GlobalFixTables[:]...)
	}
	if ok := busSec.HasKey("base_table_primary_keys"); ok {
		baseTablePrimaryKeys := strings.Split(busSec.Key("base_table_primary_keys").Value(), ",")
		for i := range baseTablePrimaryKeys {
			baseTablePrimaryKeys[i] = strings.TrimSpace(baseTablePrimaryKeys[i])
			infos := strings.Split(baseTablePrimaryKeys[i], ":")
			for j := range infos {
				infos[j] = strings.TrimSpace(infos[j])
			}
			tableName := infos[0]
			if tableName == "" {
				continue
			}
			if len(infos) > 1 {
				pks := strings.Join(infos[1:], ",")
				comm.FixTableMap[tableName] = pks
			} else if val, ok := comm.GlobalFixTableMap[tableName]; ok {
				comm.FixTableMap[tableName] = val
			}
		}
	} else {
		comm.FixTableMap = comm.GlobalFixTableMap
	}

	if ok := busSec.HasKey("blob_proto_files"); ok {

		blobProtoFiles := strings.Split(busSec.Key("blob_proto_files").Value(), ",")
		for i := range blobProtoFiles {
			blobProtoFiles[i] = strings.TrimSpace(blobProtoFiles[i])
			infos := strings.Split(blobProtoFiles[i], ":")
			for j := range infos {
				infos[j] = strings.TrimSpace(infos[j])
			}
			blobType := infos[0]
			if blobType == "" {
				continue
			}

			if len(infos) > 0 {
				filename := infos[1]
				comm.BlobFiles[blobType] = filename
			} else if val, ok := comm.GlobalBlobFiles[blobType]; ok {
				comm.BlobFiles[blobType] = val
			}
		}
	} else {
		comm.BlobFiles = comm.GlobalBlobFiles
	}
	if ok := busSec.HasKey("blob_user_in_msg_name"); ok {

		name := strings.TrimSpace(busSec.Key("blob_user_in_msg_name").Value())
		if name != "" {
			comm.BlobUserInMsg = name
		}
	}
	if ok := busSec.HasKey("blob_user_out_msg_name"); ok {

		name := strings.TrimSpace(busSec.Key("blob_user_out_msg_name").Value())
		if name != "" {
			comm.BlobUserOutMsg = name
		}
	}
	if ok := busSec.HasKey("proto_file_ignores"); ok {
		ignores := strings.TrimSpace(busSec.Key("proto_file_ignores").Value())
		if ignores != "" {
			comm.IgnoreProtoFiles = ignores
		}
	}
	if ok := busSec.HasKey("proto_file_common_name"); ok {
		name := strings.TrimSpace(busSec.Key("proto_file_common_name").Value())
		if name != "" {
			comm.CommonProtoFile = name
		}
	}
	if ok := busSec.HasKey("proto_file_enum_name"); ok {
		name := strings.TrimSpace(busSec.Key("proto_file_enum_name").Value())
		if name != "" {
			comm.EnumProtoFile = name
		}
	}
	if ok := busSec.HasKey("import_path_ignores"); ok {
		ignores := strings.Split(busSec.Key("import_path_ignores").Value(), ",")
		for i := range ignores {
			ignores[i] = strings.TrimSpace(ignores[i])
		}
		if len(ignores) > 0 {
			comm.IgnoreImportPaths = append(comm.IgnoreImportPaths, ignores[:]...)
		} else {
			comm.IgnoreImportPaths = append(comm.IgnoreImportPaths, comm.GlobalIgnoreImportPaths[:]...)
		}
	} else {
		comm.IgnoreImportPaths = append(comm.IgnoreImportPaths, comm.GlobalIgnoreImportPaths[:]...)
	}

	tcaplusSec, err := cfg.GetSection("tcaplusdb")
	if err != nil {
		fmt.Println(err)
		return err
	}
	if ok := tcaplusSec.HasKey("tcaplus_package_name"); ok {
		name := strings.TrimSpace(tcaplusSec.Key("tcaplus_package_name").Value())
		if name != "" {
			comm.TcaplusPackageName = name
		}
	}
	if ok := tcaplusSec.HasKey("tcaplus_import_path"); ok {
		name := strings.TrimSpace(tcaplusSec.Key("tcaplus_import_path").Value())
		if name != "" {
			comm.TcaplusImportName = name
		}
	}
	return nil

}
