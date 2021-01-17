package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/tencentyun/proto-parse-tcaplus/comm"
	"gopkg.in/ini.v1"
)

//read ini config file
func ReadIni(iniFile string) (*ini.File, error) {

	if _, err := os.Stat(iniFile); os.IsNotExist(err) {
		fmt.Println(iniFile + " is no exist")
		return nil, err
	}

	cfg, err := ini.Load(iniFile)
	//	fmt.Println("ini load error :" + err.Error())
	return cfg, err

}

//parse ini config file
func ParseCfg(cfg *ini.File) error {
	if cfg == nil {
		fmt.Println("cfg is nil")
	}
	busSec, err := cfg.GetSection("business")
	if err != nil {
		fmt.Println(err)
		return err
	}

	//init map object to avoid invalid address error
	comm.BaseTableMap = make(map[string]string)
	comm.BlobFiles = make(map[string]string)
	comm.TableFiles = make(map[string]string)

	if ok := busSec.HasKey("base_tables"); ok {
		//parse base tables
		baseTables := strings.Split(busSec.Key("base_tables").Value(), ",")
		for i := range baseTables {
			baseTables[i] = strings.TrimSpace(baseTables[i])
			comm.BaseTables = append(comm.BaseTables, baseTables[i])
		}

	} else {
		//init default values
		comm.BaseTables = append(comm.BaseTables, comm.GlobalBaseTables[:]...)
	}
	if ok := busSec.HasKey("base_table_primary_keys"); ok {
		//parse base table primary keys
		baseTablePrimaryKeys := strings.Split(busSec.Key("base_table_primary_keys").Value(), ",")
		for i := range baseTablePrimaryKeys {
			baseTablePrimaryKeys[i] = strings.TrimSpace(baseTablePrimaryKeys[i])
			infos := strings.Split(baseTablePrimaryKeys[i], ":")
			for j := range infos {
				//trim space to avoid illegal config item
				infos[j] = strings.TrimSpace(infos[j])
			}
			tableName := infos[0]
			if tableName == "" {
				continue
			}
			if len(infos) > 1 {
				pks := strings.Join(infos[1:], ",")
				comm.BaseTableMap[tableName] = pks
			} else if val, ok := comm.GlobalBaseTableMap[tableName]; ok {
				comm.BaseTableMap[tableName] = val
			}
		}
	} else {
		//if no config item , assign default value
		comm.BaseTableMap = comm.GlobalBaseTableMap
	}

	if ok := busSec.HasKey("table_proto_files"); ok {
		//parse config , get pub and split proto files
		protoFiles := strings.Split(busSec.Key("table_proto_files").Value(), ",")
		for i := range protoFiles {
			protoFiles[i] = strings.TrimSpace(protoFiles[i])
			infos := strings.Split(protoFiles[i], ":")
			for j := range infos {
				infos[j] = strings.TrimSpace(infos[j])
			}
			msgType := infos[0]
			if msgType == "" {
				continue
			}
			if len(infos) > 0 {
				filename := infos[1]
				comm.TableFiles[msgType] = filename
			} else if val, ok := comm.GlobalTableFiles[msgType]; ok {
				comm.TableFiles[msgType] = val
			}
		}
	} else {
		comm.TableFiles = comm.GlobalTableFiles
	}

	if ok := busSec.HasKey("blob_proto_files"); ok {
		//parse config , get blob proto files
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

	if ok := busSec.HasKey("import_path_ignores"); ok {
		ignores := strings.Split(busSec.Key("import_path_ignores").Value(), ",")
		for i := range ignores {
			ignores[i] = strings.TrimSpace(ignores[i])
		}
		if len(ignores) > 0 {
			comm.IgnoreImportPaths = append(comm.IgnoreImportPaths, ignores[:]...)
		} else {
			//empty item in config file, assign default value
			comm.IgnoreImportPaths = append(comm.IgnoreImportPaths, comm.GlobalIgnoreImportPaths[:]...)
		}
	} else {
		//no item in config file, assign default value
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
