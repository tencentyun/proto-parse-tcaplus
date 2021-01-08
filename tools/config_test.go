package tools

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tencentyun/proto-parse-tcaplus/comm"
)

func TestReadIni(t *testing.T) {
	filename := "../config/proto_parse.cfg"
	_, err := ReadIni(filename)
	assert.NoError(t, err)

	assert.NoError(t, err)
}
func TestParseCfg(t *testing.T) {
	cfg, err := ReadIni("../config/proto_parse.cfg")
	assert.NoError(t, err)

	err = ParseCfg(cfg)

	assert.NoError(t, err)
	assert.Equal(t, "BaseVersion", comm.FixTables[0])
	assert.Equal(t, "BaseGUID", comm.FixTables[1])
	assert.Equal(t, "BaseSelfIncrementIDData", comm.FixTables[2])
	assert.Equal(t, "BaseAccounts", comm.FixTables[3])
	assert.Equal(t, "BaseRoles", comm.FixTables[4])

	assert.Equal(t, "version", comm.FixTableMap["BaseVersion"])
	assert.Equal(t, "guid,uid", comm.FixTableMap["BaseGUID"])
	assert.Equal(t, "id", comm.FixTableMap["BaseSelfIncrementIDData"])
	assert.Equal(t, "token", comm.FixTableMap["BaseAccounts"])
	assert.Equal(t, "roleID", comm.FixTableMap["BaseRoles"])

	assert.Equal(t, "blob_user_data_in.proto", comm.BlobFiles["IN"])
	assert.Equal(t, "blob_user_data_out.proto", comm.BlobFiles["OUT"])

	assert.Equal(t, "blob_user_data_in", comm.BlobUserInMsg)
	assert.Equal(t, "blob_user_data_out", comm.BlobUserOutMsg)

	assert.Equal(t, "common.proto, enumm_entity.proto", comm.IgnoreProtoFiles)
	assert.Equal(t, "common.proto", comm.CommonProtoFile)
	assert.Equal(t, "enumm_entity.proto", comm.EnumProtoFile)
	assert.Equal(t, "proto/entity/common.proto,proto/entity/enumm_entity.proto", strings.Join(comm.IgnoreImportPaths, ","))
	assert.Equal(t, "tcaplus_entity", comm.TcaplusPackageName)
	assert.Equal(t, "tcaplusservice.optionv1.proto", comm.TcaplusImportName)

}
