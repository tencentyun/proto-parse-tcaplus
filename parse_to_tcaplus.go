package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
	"github.com/tencentyun/proto-parse-tcaplus/comm"
	"github.com/tencentyun/proto-parse-tcaplus/tools"
)

type ProtoInfo struct {
	enums  []comm.Enum
	msgs   []comm.Message
	imps   []comm.Import
	pkg    comm.Package
	syntax comm.Syntax
	opts   []comm.Option
}

//global variables
var (
	buf bytes.Buffer
	//struct object for parsing
	protoInfo ProtoInfo
	//struct object for parsing
	protoInfos = map[string]ProtoInfo{}

	//save errors for each proto file
	errorInfos = map[string]string{}

	//save base messages
	baseMessages []comm.Message
	//save blob IN and OUT messages
	blobMessages = map[string][]string{}
	//split messages, message with IN or OUT prefix, UUID:primary key, UID: index
	splitMessages []comm.Message
	//pub messages, message with PUB prefix, UUID: primary key
	pubMessages []comm.Message

	//save other messages (not  base, blob, split, and pub)
	commMessages []comm.Message
	//save all enums
	commEnums []comm.Enum

	//temp variable for enum field, key: msgtype, value: enum list
	tempEnums = map[string][]comm.Enum{}

	//package name
	GeneralPackageName string = "entity"
)

//parse proto file and generate new proto file
func ProtoParseAndWrite(srcPath string, dstPath string, ignores string) {
	//traverse all proto files and parse them
	err := traverseProtoFiles(srcPath, ignores)
	if err != nil {
		fmt.Println(err)
		return
	}
	//classify message type
	err = classifyProtoFiles(srcPath, ignores, dstPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	//generate proto files with parsed results
	writeProtoFiles(dstPath)

	//output parse results for each proto file, SUCCESS or FAIL
	err = outputParseResults(srcPath, ignores)
	if err != nil {
		fmt.Println(err)
		return
	}

}

/*
* @param srcPath string : source path of proto files
* @param ignores string : specify the proto files need to be ignored, comma sepeartes each proto file
* @retval error
 */
func traverseProtoFiles(srcPath string, ignores string) error {
	//get all proto files from source path, ignoring the proto files specfied by ignores
	protoFiles, err := tools.GetProtoFiles(srcPath, ignores)
	if err != nil {
		return fmt.Errorf("get proto files error : %v", err)

	}
	//loop for proto files
	for _, file := range protoFiles {
		filename := path.Base(file)
		//parse proto file and save results into protoInfo (global variable)
		parse(file)
		//add additional contents to protoInfo
		protoInfo.imps = append(protoInfo.imps, comm.Import{Path: comm.TcaplusImportName})
		//map the protoInfo to relative proto file , and save  into protoInfos
		//user can scan all parsed results of proto file from protoInfos with proto file name
		protoInfos[filename] = protoInfo
		//reset protoInfo for next proto file
		protoInfo = ProtoInfo{}
	}
	return nil
}

/*
* @brief check the message type of parse results, and separate them into different entities, such base entity, blob entity, split entity (in and out), and pub entity
 */
func classifyProtoFiles(srcPath string, ignores string, dstPath string) error {
	protoFiles, err := tools.GetProtoFiles(srcPath, ignores)
	if err != nil {
		return err
	}
	for _, file := range protoFiles {
		filename := path.Base(file)
		info, ok := protoInfos[filename]
		if !ok {
			return fmt.Errorf("%s no parse results.", filename)
		}
		for _, msg := range info.msgs {
			//newName := tools.SnakeCase(msg.Name)
			if blobType, ok := isBlobMessageType(msg); ok {
				blobMessages[blobType] = append(blobMessages[blobType], msg.Name)
			} else if _, ok := isInOrOutMessageType(msg); ok {
				splitMessages = append(splitMessages, msg)
			} else if _, ok := isPubMessageType(msg); ok {
				pubMessages = append(pubMessages, msg)
			} else if ok := isBaseMessageType(msg); ok {
				baseMessages = append(baseMessages, msg)
			} else {
				commMessages = append(commMessages, msg)
			}

		}
		for _, e := range info.enums {
			checkAndAppendCommEnums(e)
		}
	}
	return nil
}

//check the enum duplication in commEnums
func checkAndAppendCommEnums(e comm.Enum) {
	existFlag := 0
	for _, ee := range commEnums {
		if e.Name == ee.Name {
			existFlag = 1
			break
		}
	}
	if existFlag == 0 {
		commEnums = append(commEnums, e)
	}
}

//generate proto files, ignore generating common.proto and enumm_entity.proto
func writeProtoFiles(dstPath string) {
	writeBaseProtoFiles(dstPath)
	writeBlobProtoFiles(dstPath)
	writeSplitProtoFiles(dstPath)
	writePubProtoFiles(dstPath)

}

//output results for checking whether the parsing is ok or not
func outputParseResults(srcPath string, ignores string) error {
	protoFiles := []string{comm.TableFiles["BASE"], comm.TableFiles["PUB"], comm.TableFiles["SPLIT"], comm.BlobFiles["IN"], comm.BlobFiles["OUT"]}
	for _, file := range protoFiles {
		filename := path.Base(file)
		if err, ok := errorInfos[filename]; ok {
			fmt.Println(fmt.Sprintf("[%v] convert [FAIL][%v]", filename, err))
		} else {
			fmt.Println(fmt.Sprintf("[%v] convert [SUCCESS]", filename))
		}
	}
	return nil
}

//parse proto file
func parse(protoSrcFile string) {

	reader, _ := os.Open(protoSrcFile)
	defer reader.Close()
	//parse the proto syntax tree
	parser := proto.NewParser(reader)
	definition, _ := parser.Parse()
	//walk the proto file
	proto.Walk(definition,
		protoWithSyntax(handleSyntax),
		proto.WithImport(handleImport),
		proto.WithPackage(handlePackage),
		proto.WithOption(handleOption),
		proto.WithEnum(handleEnum),
		proto.WithMessage(handleMessage),
	)

}

func protoWithSyntax(apply func(p *proto.Syntax)) proto.Handler {
	return func(v proto.Visitee) {
		if s, ok := v.(*proto.Syntax); ok {
			apply(s)
		}
	}
}
func handleSyntax(s *proto.Syntax) {
	protoInfo.syntax.Name = s.Value
}
func handleImport(im *proto.Import) {
	//ignore general imports

	imp := comm.Import{
		Path: im.Filename,
	}
	for _, ignorePath := range comm.IgnoreImportPaths {
		if im.Filename == ignorePath {
			return
		}
	}
	protoInfo.imps = append(protoInfo.imps, imp)
}

func handlePackage(p *proto.Package) {
	protoInfo.pkg = comm.Package{
		Name: comm.TcaplusPackageName,
	}
}

func handleOption(o *proto.Option) {
	if _, ok := o.Parent.(*proto.Proto); !ok {
		//skip the nested option in message
		return
	}
	//not parse option of business proto, meaningless for tcaplusdb
	//ToDO
}

func handleEnum(e *proto.Enum) {
	/*
		if p, ok := e.Parent.(*proto.Message); ok {
			if p != nil {
				e.Name = fmt.Sprintf("%s.%s", p.Name, e.Name)
			}
		}
	*/
	if _, ok := e.Parent.(*proto.Proto); !ok {
		//skip the enum defined in message
		return
	}
	protoInfo.enums = append(protoInfo.enums, parseEnum(e))

}
func parseEnum(e *proto.Enum) comm.Enum {
	enum := comm.Enum{
		Name: e.Name,
	}

	for _, v := range e.Elements {
		//handle enum option

		if _, ok := v.(*proto.Option); ok {
			//not parse, meaningless for tcaplusdb
		}

		//handle enum field
		if ef, ok := v.(*proto.EnumField); ok {

			field := comm.EnumField{
				Name:    ef.Name,
				Integer: ef.Integer,
			}
			enum.EnumFields = append(enum.EnumFields, field)
		}

	}

	return enum
}
func handleMessage(m *proto.Message) {
	if _, ok := m.Parent.(*proto.Proto); !ok {
		//if the message is nested
		return
	}
	protoInfo.msgs = append(protoInfo.msgs, parseMessage(m))
}
func parseMessage(m *proto.Message) comm.Message {
	msg := comm.Message{
		Name: m.Name,
	}
	for _, v := range m.Elements {
		if _, ok := v.(*proto.Option); ok {
			//not parse, meaningless for tcaplusdb
		}
		if f, ok := v.(*proto.NormalField); ok {
			msg.Fields = append(msg.Fields, comm.Field{
				ID:         f.Sequence,
				Name:       f.Name,
				Type:       f.Type,
				IsRepeated: f.Repeated,
			})
		}
		if mmp, ok := v.(*proto.MapField); ok {
			f := mmp.Field
			msg.Maps = append(msg.Maps, comm.Map{
				KeyType: mmp.KeyType,
				Field: comm.Field{
					ID:         f.Sequence,
					Name:       f.Name,
					Type:       f.Type,
					IsRepeated: false,
				},
			})
		}

		if moo, ok := v.(*proto.Oneof); ok {
			var fields []comm.Field
			for _, el := range moo.Elements {
				if f, ok := el.(*proto.OneOfField); ok {
					fields = append(fields, comm.Field{
						ID:         f.Sequence,
						Name:       f.Name,
						Type:       f.Type,
						IsRepeated: false,
					})
				}
			}
			msg.Fields = append(msg.Fields, fields...)
		}

		if _, ok := v.(*proto.Reserved); ok {
			//not parse
		}

		if m, ok := v.(*proto.Message); ok {
			msg.Messages = append(msg.Messages, parseMessage(m))
		}
		if e, ok := v.(*proto.Enum); ok {
			msg.Enums = append(msg.Enums, parseEnum(e))
		}
	}
	return msg
}

func writeBaseProtoFiles(dstPath string) {
	errStr := ""
	baseProtoFileName := comm.TableFiles["BASE"]
	dstFile := filepath.Join(dstPath, baseProtoFileName)

	//write syntax, package, import
	writeProtoFileHead()
	for _, msg := range baseMessages {
		err := writeBaseMessage(msg)
		if err != nil {
			errStr = fmt.Sprintf("%s;%s", errStr, err.Error())
		}
	}
	/*
		//write nested enums
		if es, ok := tempEnums["BASE"]; ok {
			for _, e := range es {
				writeEnum(e)
			}
		}
	*/
	err := tools.WriteFile(dstFile, buf.Bytes())
	if err != nil {
		errStr = fmt.Sprintf("%s;%s", errStr, err.Error())
	}
	buf.Reset()

	if errStr != "" {
		errorInfos[baseProtoFileName] = errStr
	}
}

//put parse results into bytes.Buffer
func writeSplitProtoFiles(dstPath string) {
	errStr := ""
	splitProtoFileName := comm.TableFiles["SPLIT"]
	dstFile := filepath.Join(dstPath, splitProtoFileName)
	//write syntax, package, import
	writeProtoFileHead()

	for _, msg := range splitMessages {
		err := writeSplitMessage(msg, "SPLIT")
		if err != nil {
			errStr = fmt.Sprintf("%s;%s", errStr, err.Error())
		}
	}
	/*
		//write nested enums
		if es, ok := tempEnums["SPLIT"]; ok {
			for _, e := range es {
				writeEnum(e)
			}
		}
	*/
	err := tools.WriteFile(dstFile, buf.Bytes())
	if err != nil {
		errStr = fmt.Sprintf("%s;%s", errStr, err.Error())
	}
	buf.Reset()

	if errStr != "" {
		errorInfos[splitProtoFileName] = errStr
	}

}

func writePubProtoFiles(dstPath string) {
	errStr := ""
	pubProtoFileName := comm.TableFiles["PUB"]
	dstFile := filepath.Join(dstPath, pubProtoFileName)
	//write syntax, package, import
	writeProtoFileHead()
	for _, msg := range pubMessages {
		err := writePubMessage(msg, "PUB")
		if err != nil {
			errStr = fmt.Sprintf("%s;%s", errStr, err.Error())
		}
	}
	/*
		//write nested enums
		if es, ok := tempEnums["PUB"]; ok {
			for _, e := range es {
				writeEnum(e)
			}
		}
	*/
	err := tools.WriteFile(dstFile, buf.Bytes())
	if err != nil {
		errStr = fmt.Sprintf("%s;%s", errStr, err.Error())
	}
	buf.Reset()

	if errStr != "" {
		errorInfos[pubProtoFileName] = errStr
	}
}

func writeBlobProtoFiles(dstPath string) {
	//write BLOB messages to specified message (blob_user_data_out, blob_user_data_in)
	for msgType, file := range comm.BlobFiles {
		msgs, ok := blobMessages[msgType]
		if !ok {
			errorInfos[file] = fmt.Sprintf("no %v blob messages", msgType)
		} else {
			err := writeBlobMessages(msgType, msgs)
			if err != nil {
				errorInfos[file] = err.Error()
			}
			dstFile := filepath.Join(dstPath, file)
			err = tools.WriteFile(dstFile, buf.Bytes())
			if err != nil {
				errorInfos[file] = err.Error()
			}
		}

		//reset to empty for next proto file
		buf.Reset()
	}
}
func writeProtoFileHead() {
	buf.WriteString("syntax = \"proto3\";\n")
	buf.WriteString(fmt.Sprintf("package %v;\n", comm.TcaplusPackageName))
	buf.WriteString(fmt.Sprintf("import \"%s\";\n", comm.TcaplusImportName))
}

func writeImports(info ProtoInfo) {
	if len(info.imps) == 0 {
		//fmt.Println("no import need to be written")
		return
	}
	for _, i := range info.imps {
		buf.WriteString(fmt.Sprintf("import \"%v\";\n", i.Path))
	}
}

func writeEnums(info ProtoInfo) {
	if len(info.enums) == 0 {
		//fmt.Println("no enum need to be written")
		return
	}

	for _, e := range info.enums {
		writeEnum(e)
	}

}

func writeEnum(e comm.Enum) {
	buf.WriteString(fmt.Sprintf("enum %s {\n", e.Name))
	for _, field := range e.EnumFields {
		buf.WriteString(fmt.Sprintf("\t%v = %v;\n", field.Name, field.Integer))
	}
	buf.WriteString("}\n")
}

func writeBaseMessage(msg comm.Message) error {
	if pk, ok := comm.BaseTableMap[msg.Name]; ok {
		//	newName := tools.SnakeCase(msg.Name)
		buf.WriteString(fmt.Sprintf("message %s{\n", msg.Name))
		optStr := fmt.Sprintf("\toption(tcaplusservice.tcaplus_primary_key) = \"%s\";\n", pk)
		buf.WriteString(optStr)
	} else {
		return fmt.Errorf("write %s message option error, message name not in BaseTableMap", msg.Name)
	}
	if err := writeMessageBody(msg, "BASE"); err != nil {
		return err
	}
	buf.WriteString("}\n")
	return nil
}
func writeSplitMessage(msg comm.Message, msgType string) error {
	// newName := tools.SnakeCase(msg.Name)
	buf.WriteString(fmt.Sprintf("message %s{\n", msg.Name))
	optStr := fmt.Sprintf("\toption(tcaplusservice.tcaplus_primary_key) = \"UUID,UID\";\n")
	buf.WriteString(optStr)
	optStr = fmt.Sprintf("\toption(tcaplusservice.tcaplus_index) = \"index_1(UID)\";\n")
	buf.WriteString(optStr)
	if err := writeMessageBody(msg, msgType); err != nil {
		return err
	}
	buf.WriteString("}\n")
	return nil
}
func writePubMessage(msg comm.Message, msgType string) error {
	//newName := tools.SnakeCase(msg.Name)
	buf.WriteString(fmt.Sprintf("message %s{\n", msg.Name))
	optStr := fmt.Sprintf("\toption(tcaplusservice.tcaplus_primary_key) = \"UUID\";\n")
	buf.WriteString(optStr)
	if err := writeMessageBody(msg, msgType); err != nil {
		return err
	}
	buf.WriteString("}\n")
	return nil
}
func writeBlobMessages(msgType string, msgs []string) error {
	writeProtoFileHead()
	if msgType == "OUT" {
		buf.WriteString(fmt.Sprintf("message %v { \n", comm.BlobUserOutMsg))
		optStr := fmt.Sprintf("\toption(tcaplusservice.tcaplus_primary_key) = \"UID\";\n")
		buf.WriteString(optStr)
		buf.WriteString(fmt.Sprintf("\tuint64 UID = 1;\n\tuint64 UpdateTime = 2;\n"))
		seqId := 3
		for _, oms := range msgs {
			buf.WriteString(fmt.Sprintf("\tbytes %v = %d;\n", oms, seqId))
			seqId = seqId + 1
		}
		buf.WriteString("}\n")
	} else if msgType == "IN" {
		buf.WriteString(fmt.Sprintf("message %v { \n", comm.BlobUserInMsg))
		optStr := fmt.Sprintf("\toption(tcaplusservice.tcaplus_primary_key) = \"UID\";\n")
		buf.WriteString(optStr)
		buf.WriteString(fmt.Sprintf("\tuint64 UID = 1;\n\tuint64 UpdateTime = 2;\n"))
		seqId := 3
		for _, ims := range msgs {
			buf.WriteString(fmt.Sprintf("\tbytes %v = %d;\n", ims, seqId))
			seqId = seqId + 1
		}
		buf.WriteString("}\n")
	}
	return nil
}

func writeMessageBody(msg comm.Message, msgType string) error {
	seqIncr := 0
	maxSeq := 0
	for _, field := range msg.Fields {
		fieldStr := ""

		if msgType == "BASE" {
			maxSeq = field.ID
		}

		if field.Type == "EntityType" {
			if msgType == "BASE" {
				//if message is base message, the start sequence id need decrease 1 because of getting rid of EntityType field
				seqIncr = -1
			}
			//skip EntityType field
			continue
		}
		if field.Name == "UUID" && (msgType == "SPLIT") {
			fieldStr = fmt.Sprintf("\t%v %v = 1;\n\tuint64 UID = 2;\n\tuint64 UpdateTime = 3;\n", field.Type, field.Name)
			buf.WriteString(fieldStr)
			seqIncr = 1 //increase 1
			continue
		}
		if field.Name == "UUID" && msgType == "PUB" {
			fieldStr = fmt.Sprintf("\t%v %v = 1;\n\tuint64 UpdateTime = 2;\n", field.Type, field.Name)
			buf.WriteString(fieldStr)
			continue
		}
		if field.IsRepeated {
			fieldStr = "repeated "
		}
		newId := field.ID + seqIncr
		newName := strings.Title(field.Name)
		if _, ok := isEnumInCommEnums(field.Type); ok {
			//enum field, nested enums or defined in common proto file (enumm_entity.proto)
			//convert all enums to int32
			fieldStr = fmt.Sprintf("\t%vint32 %v = %v;\n", fieldStr, newName, newId)
			//add enum into temp list
			//checkAndAppendTempEnums(msgType, *e)
		} else if ok := isNestedEnum(field.Type, msg); ok {
			fieldStr = fmt.Sprintf("\t%vint32 %v = %v;\n", fieldStr, newName, newId)
		} else if ok := isMessageInCommMessages(field.Type); ok {
			//message (not base, pub, split, and blob message)
			fieldStr = fmt.Sprintf("\t%vbytes %v = %v;\n", fieldStr, newName, newId)
		} else if ok := isNestedMessage(field.Type, msg); ok {
			//nested message field, defined in current message, convert to bytes
			fieldStr = fmt.Sprintf("\t%vbytes %v = %v;\n", fieldStr, newName, newId)
		} else {
			fieldStr = fmt.Sprintf("\t%v%v %v = %v;\n", fieldStr, field.Type, newName, newId)
		}

		buf.WriteString(fieldStr)
	}

	//deal with base table rules
	if msgType == "BASE" && maxSeq != 0 {
		if msg.Name == "BaseAccounts" {
			buf.WriteString(fmt.Sprintf("\tuint64 AddTime = %d;\n\tuint64 UpdateTime = %d;\n", maxSeq, maxSeq+1))
		} else {
			buf.WriteString(fmt.Sprintf("\tuint64 UpdateTime = %d;\n", maxSeq))
		}

	}

	for _, mapf := range msg.Maps {
		newId := mapf.Field.ID + seqIncr
		newName := strings.Title(mapf.Field.Name)
		buf.WriteString(fmt.Sprintf("\tbytes %v = %v;\n", newName, newId))
	}
	for _, enumf := range msg.Enums {
		//deal nested enums
		writeEnum(enumf)
	}

	/*
			for _, msgf := range msg.Messages {
		        //not deal, nested message will be converted to bytes,
			}
	*/
	return nil
}
func checkAndAppendTempEnums(msgType string, e comm.Enum) {
	existFlag := 0
	if es, ok := tempEnums[msgType]; ok {
		for _, ee := range es {
			if e.Name == ee.Name {
				existFlag = 1
				break
			}
		}
	}
	if existFlag == 0 {
		tempEnums[msgType] = append(tempEnums[msgType], e)
	}
}
func isEnumInCommEnums(name string) (*comm.Enum, bool) {
	replaceStr := fmt.Sprintf("%s.", GeneralPackageName)
	for _, e := range commEnums {
		if name == e.Name {
			return &e, true
		}
		newName := strings.TrimPrefix(name, replaceStr)
		if newName == e.Name {
			return nil, true
		}
	}
	return nil, false
}
func isMessageInCommMessages(name string) bool {
	replaceStr := fmt.Sprintf("%s.", GeneralPackageName)
	for _, m := range commMessages {
		if name == m.Name {
			return true
		}
		//some field is message type with package prefix, such as: entity.WORD_POS postion=1;
		newName := strings.TrimPrefix(name, replaceStr)
		if newName == m.Name {
			return true
		}

	}

	return false
}
func isNestedMessage(name string, msg comm.Message) bool {
	//message is nested in current message
	for _, m := range msg.Messages {
		if name == m.Name {
			return true
		}
	}
	return false
}
func isNestedEnum(name string, msg comm.Message) bool {
	//enum is nested in current message
	for _, e := range msg.Enums {
		if name == e.Name {
			return true
		}
	}
	return false
}

func isBaseMessageType(msg comm.Message) bool {
	//check base type (such account, role,etc.)
	//check base type
	for _, bs := range comm.BaseTables {
		if msg.Name == bs {
			return true
		}
	}
	return false
}
func isBlobMessageType(msg comm.Message) (string, bool) {
	//check blob message type, message feature: OUT prefix or IN prefix , only has EntityType field without UUID field
	//message will be added to blob_user_data_out (message with OUT prefix) or blob_user_data_in (message with IN prefix) message
	// blob message will be converted to bytes type and be  generated to tcaplusdb table
	blobType := ""
	if strings.HasPrefix(msg.Name, "OUT_") {
		blobType = "OUT"
	} else if strings.HasPrefix(msg.Name, "IN_") {
		blobType = "IN"
	}
	flag := checkMessageFlag(msg)
	if blobType != "" && flag == 1 {
		//is blob message
		return blobType, true
	}
	return "", false
}
func isInOrOutMessageType(msg comm.Message) (string, bool) {
	//check in or out message, message feature: IN_ or OUT_ prefix, both EntityType and UUID exist
	//message will be generated to tcaplusdb table
	msgType := ""
	if strings.HasPrefix(msg.Name, "OUT_") {
		msgType = "OUT"
	} else if strings.HasPrefix(msg.Name, "IN_") {
		msgType = "IN"
	}
	flag := checkMessageFlag(msg)
	if msgType != "" && flag == 2 {
		return msgType, true
	}
	return "", false
}
func isPubMessageType(msg comm.Message) (string, bool) {
	//check pub message, message feature: PUB prefix, both EntityType and UUID exist
	//message will be generated to tcaplusdb table
	flag := checkMessageFlag(msg)
	if strings.HasPrefix(msg.Name, "PUB_") && flag == 2 {
		return "PUB", true
	}
	return "", false
}

func checkMessageFlag(msg comm.Message) int {
	flag := 0
	for _, field := range msg.Fields {
		if field.Type == "EntityType" {
			flag = 1
			continue
		}
		if field.Name == "UUID" {
			flag = flag + 1
			break
		}
	}
	return flag
}
