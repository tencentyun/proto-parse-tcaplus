package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
	"github.com/tencentyun/proto-parse-tcaplus/tools"
)

//global variables
//      "enums": []Enum,
//		"msgs":   []Message,
//		"imps":   []Import,
//		"pkg":    Package,
//		"syntax": Syntax,
//		"opts":  []Option,

type ProtoInfo struct {
	enums  []Enum
	msgs   []Message
	imps   []Import
	pkg    Package
	syntax Syntax
	opts   []Option
}

//global variables
var (
	buf bytes.Buffer
	//struct object for parsing
	protoInfo ProtoInfo
	//struct object for parsing
	protoInfos = map[string]ProtoInfo{}
	//save parsed results of common.proto
	commonProtoInfo ProtoInfo
	//save parsed results of enumm_entity.proto
	enummProtoInfo ProtoInfo
	//save temp enum results for writing
	tempEnumInfos []Enum
	//save temp message results for writing
	tempMsgInfos []Message
	//save errors for each proto file
	errorInfos   = map[string]string{}
	blobMessages = map[string][]string{}
)

func ProtoParseAndWrite(srcPath string, dstPath string, ignores string) {
	err := traverseProtoFiles(srcPath, ignores)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = writeProtoFiles(srcPath, ignores, dstPath)
	if err != nil {
		fmt.Println(err)
		return
	}
}
func traverseProtoFiles(srcPath string, ignores string) error {
	protoFiles, err := tools.GetProtoFiles(srcPath, ignores)
	if err != nil {
		return fmt.Errorf("get proto files error : %v", err)

	}

	for _, file := range protoFiles {
		filename := path.Base(file)
		//parse proto file and save results into protoInfo (global variable)
		parse(file)
		//add additional contents to protoInfo
		protoInfo.imps = append(protoInfo.imps, Import{Path: TcaplusImportName})
		//map the protoInfo to relative proto file , and save  into protoInfos
		//user can scan all parsed results of proto file from protoInfos with proto file name
		protoInfos[filename] = protoInfo
		//reset protoInfo for next proto file
		protoInfo = ProtoInfo{}
	}
	//parse common proto, common.proto, enumm_entity.proto

	for _, filename := range strings.Split(ignores, ",") {
		commfile := filepath.Join(srcPath, filename)
		parse(commfile)
		//add additional contents to protoInfo
		protoInfo.imps = append(protoInfo.imps, Import{Path: TcaplusImportName})
		//map the protoInfo to relative proto file , and save  into protoInfos
		//user can scan all parsed results of proto file from protoInfos with proto file name
		protoInfos[filename] = protoInfo
		//reset protoInfo for next proto file
		protoInfo = ProtoInfo{}
	}
	if protoInfo, ok := protoInfos[CommonProtoFile]; ok {
		commonProtoInfo = protoInfo
	} else {
		return fmt.Errorf("parse common.proto fail, please check")
	}
	if enumInfo, ok := protoInfos[EnumProtoFile]; ok {
		enummProtoInfo = enumInfo
	} else {
		return fmt.Errorf("parse enumm_entity.proto fail, please check")
	}
	return nil
}

func writeProtoFiles(srcPath string, ignores string, dstPath string) error {
	protoFiles, err := tools.GetProtoFiles(srcPath, ignores)
	if err != nil {
		return err
	}
	for _, file := range protoFiles {
		filename := path.Base(file)
		dstFile := filepath.Join(dstPath, filename)
		writeProtoFile(dstFile)
		err := tools.WriteFile(dstFile, buf.Bytes())
		if err != nil {
			return err
		}
		//reset to empty for next proto file
		buf.Reset()
	}
	outputParseResults(protoFiles)
	return nil
}

func outputParseResults(protoFiles []string) {
	for _, file := range protoFiles {
		filename := path.Base(file)
		if err, ok := errorInfos[filename]; ok {
			fmt.Println(fmt.Sprintf("[%v] convert [FAIL][%v]\n", filename, err))
		} else {
			fmt.Println(fmt.Sprintf("[%v] convert [SUCCESS]", filename))
		}
	}
}

func parse(protoSrcFile string) {

	reader, _ := os.Open(protoSrcFile)
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, _ := parser.Parse()

	proto.Walk(definition,
		protoWithSyntax(handleSyntax),
		proto.WithImport(handleImport),
		proto.WithPackage(handlePackage),
		proto.WithOption(handleOption),
		proto.WithEnum(handleEnum),
		proto.WithMessage(handleMessage),
	)

}

func writeProtoFile(protoDstPath string) {
	filename := path.Base(protoDstPath)
	info := protoInfos[filename]
	//write syntax
	buf.WriteString(fmt.Sprintf("syntax = %s;\n", info.syntax.Name))
	//write package
	buf.WriteString(fmt.Sprintf("package  %s;\n", info.pkg.Name))

	//skip original import of business's definition, add tcaplusdb import : "tcaplusservice.optionv1.proto"
	writeImports(info)

	//write enum
	writeEnums(info)
	//write message, distinguish different message type, BLOB, IN, OUT and PUB
	//BLOB: save to map object first , then call writeBlobMessages to write file
	//IN, OUT, PUB:  write file directly
	writeMessages(info)
	//write BLOB messages to specified message (blob_user_data_out, blob_user_data_in)
	err := writeBlobMessages(blobMessages)
	if err != nil {
		errorInfos[filename] = err.Error()
	}

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
	imp := Import{
		Path: im.Filename,
	}
	protoInfo.imps = append(protoInfo.imps, imp)

}

func handlePackage(p *proto.Package) {
	protoInfo.pkg = Package{
		Name: TcaplusPackageName,
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
func parseEnum(e *proto.Enum) Enum {
	enum := Enum{
		Name: e.Name,
	}

	for _, v := range e.Elements {
		//handle enum option

		if _, ok := v.(*proto.Option); ok {
			//not parse, meaningless for tcaplusdb
		}

		//handle enum field
		if ef, ok := v.(*proto.EnumField); ok {

			field := EnumField{
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
func parseMessage(m *proto.Message) Message {
	msg := Message{
		Name: m.Name,
	}
	for _, v := range m.Elements {
		if _, ok := v.(*proto.Option); ok {
			//not parse, meaningless for tcaplusdb
		}
		if f, ok := v.(*proto.NormalField); ok {
			msg.Fields = append(msg.Fields, Field{
				ID:         f.Sequence,
				Name:       f.Name,
				Type:       f.Type,
				IsRepeated: f.Repeated,
			})
		}
		if mmp, ok := v.(*proto.MapField); ok {
			f := mmp.Field
			msg.Maps = append(msg.Maps, Map{
				KeyType: mmp.KeyType,
				Field: Field{
					ID:         f.Sequence,
					Name:       f.Name,
					Type:       f.Type,
					IsRepeated: false,
				},
			})
		}

		if moo, ok := v.(*proto.Oneof); ok {
			var fields []Field
			for _, el := range moo.Elements {
				if f, ok := el.(*proto.OneOfField); ok {
					fields = append(fields, Field{
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

func writeImports(info ProtoInfo) {
	if len(info.imps) == 0 {
		fmt.Println("no import need to be written")
		return
	}
	for _, i := range info.imps {
		buf.WriteString(fmt.Sprintf("import \"%v\";\n", i.Path))
	}
}

func writeEnums(info ProtoInfo) {
	if len(info.enums) == 0 {
		fmt.Println("no enum need to be written")
		return
	}

	for _, e := range info.enums {
		writeEnum(e)
	}

}
func writeEnum(e Enum) {
	buf.WriteString(fmt.Sprintf("enum %s {\n", e.Name))
	for _, field := range e.EnumFields {
		buf.WriteString(fmt.Sprintf("\t%v = %v;\n", field.Name, field.Integer))
	}
	buf.WriteString("}\n")
}

func writeMessages(info ProtoInfo) error {
	if len(info.msgs) == 0 {
		fmt.Println("no message need to be written")
		return nil
	}

	for _, msg := range info.msgs {
		newName := tools.SnakeCase(msg.Name)
		if blobType, ok := isBlobMessageType(msg); ok {
			blobMessages[blobType] = append(blobMessages[blobType], newName)
		} else {
			err := writeMessage(msg, info)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func writeMessage(m Message, info ProtoInfo) error {
	msgType := ""
	newName := tools.SnakeCase(m.Name)
	buf.WriteString(fmt.Sprintf("message %s{\n", newName))
	if ok := isBaseMessageType(m); ok {
		msgType = "BASE"
		err := writeBaseMessage(m, msgType, info)
		if err != nil {
			return err
		}
	} else if msgType, ok := isInOrOutMessageType(m); ok {
		err := writeInOrOutMessage(m, msgType, info)
		if err != nil {
			return err
		}
	} else if msgType, ok := isPubMessageType(m); ok {
		err := writePubMessage(m, msgType, info)
		if err != nil {
			return err
		}
	}
	buf.WriteString("}\n")
	return nil
}

func writeBaseMessage(msg Message, msgType string, info ProtoInfo) error {
	if pk, ok := FixTableMap[msg.Name]; ok {
		optStr := fmt.Sprintf("\toption(tcaplusservice.tcaplus_primary_key) = \"%s\";\n", pk)
		buf.WriteString(optStr)
	} else {
		return fmt.Errorf("write %s message option error, message name not in FixTableMap", msg.Name)
	}
	if err := writeMessageBody(msg, msgType, info); err != nil {
		return err
	}
	return nil
}
func writeInOrOutMessage(msg Message, msgType string, info ProtoInfo) error {
	optStr := fmt.Sprintf("\toption(tcaplusservice.tcaplus_primary_key) = \"uuid\";\n")
	buf.WriteString(optStr)
	optStr = fmt.Sprintf("\toption(tcaplusservice.tcaplus_index) = \"index_1(uid)\";\n")
	buf.WriteString(optStr)
	if err := writeMessageBody(msg, msgType, info); err != nil {
		return err
	}
	return nil
}
func writePubMessage(msg Message, msgType string, info ProtoInfo) error {
	optStr := fmt.Sprintf("\toption(tcaplusservice.tcaplus_primary_key) = \"uuid\";\n")
	buf.WriteString(optStr)
	if err := writeMessageBody(msg, msgType, info); err != nil {
		return err
	}
	return nil
}
func writeBlobMessages(msgs map[string][]string) error {
	if outMsgs, ok := msgs["OUT"]; ok {
		buf.WriteString(fmt.Sprintf("message %v { \n", BlobUserOutMsg))
		optStr := fmt.Sprintf("\toption(tcaplusservice.tcaplus_primary_key) = \"uid\";\n")
		buf.WriteString(optStr)
		buf.WriteString(fmt.Sprintf("\tuint64 uid = 1;\n\tuint64 update_time = 2;\n"))
		seqId := 3
		for _, oms := range outMsgs {
			buf.WriteString(fmt.Sprintf("\tbytes %v = %d;\n", oms, seqId))
			seqId = seqId + 1
		}
		buf.WriteString("}\n")
	} else if inMsgs, ok := msgs["IN"]; ok {
		buf.WriteString(fmt.Sprintf("message %v { \n", BlobUserInMsg))
		optStr := fmt.Sprintf("\toption(tcaplusservice.tcaplus_primary_key) = \"uid\";\n")
		buf.WriteString(optStr)
		buf.WriteString(fmt.Sprintf("\tuint64 uid = 1;\n\tuint64 update_time = 2;\n"))
		seqId := 3
		for _, ims := range inMsgs {
			buf.WriteString(fmt.Sprintf("\tbytes %v = %d;\n", ims, seqId))
			seqId = seqId + 1
		}
		buf.WriteString("}\n")
	}
	return nil
}

func writeMessageBody(msg Message, msgType string, info ProtoInfo) error {
	seqIncr := 0
	for _, field := range msg.Fields {
		fieldStr := ""

		if field.Type == "EntityType" {
			//skip EntityType field
			continue
		}
		if field.Name == "UUID" && (msgType == "IN" || msgType == "OUT") {
			//skip primary key , deal with it separately
			newName := tools.SnakeCase(field.Name)
			fieldStr = fmt.Sprintf("\t%v %v = 1;\n\tuint64 uid = 2;\n\tuint64 update_time = 3;\n", field.Type, newName)
			buf.WriteString(fieldStr)
			seqIncr = 1 //increase 1
			continue
		}
		if field.Name == "UUID" && msgType == "PUB" {
			newName := tools.SnakeCase(field.Name)
			fieldStr = fmt.Sprintf("\t%v %v = 1;\n\tuint64 update_time = 2;\n", field.Type, newName)
			buf.WriteString(fieldStr)
			continue
		}
		if field.IsRepeated {
			fieldStr = "repeated "
		}
		if e, ok := isEnumInCommProtoFile(field.Type); ok {
			//enum field, defined  in common.proto or enumm_entity.proto, save this enum into temp Enum slice
			newId := field.ID + seqIncr
			fieldStr = fmt.Sprintf("\t%v%v %v = %v;\n", fieldStr, field.Type, field.Name, newId)
			tempEnumInfos = append(tempEnumInfos, *e)
		} else if ok := isMessageInCommProtoFile(field.Type); ok {
			//message field , defined in common.proto or enumm_entity.proto, convert to bytes
			newId := field.ID + seqIncr
			fieldStr = fmt.Sprintf("\t%vbytes %v = %v;\n", fieldStr, field.Name, newId)
		} else if ok := isMessageInLocalProtoFile(field.Type, info); ok {
			//message field, defined in current proto file, convert to bytes
			newId := field.ID + seqIncr
			fieldStr = fmt.Sprintf("\t%vbytes %v = %v;\n", fieldStr, field.Name, newId)
		} else if ok := isNestedMessage(field.Type, msg); ok {
			//nested message field, defined in current message, convert to bytes
			newId := field.ID + seqIncr
			fieldStr = fmt.Sprintf("\t%vbytes %v = %v;\n", fieldStr, field.Name, newId)
		}

		buf.WriteString(fieldStr)
	}

	for _, mapf := range msg.Maps {
		newId := mapf.Field.ID + seqIncr
		buf.WriteString(fmt.Sprintf("\tbytes %v = %v;\n", mapf.Field.Name, newId))
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
func isEnumInCommProtoFile(name string) (*Enum, bool) {
	for _, e := range commonProtoInfo.enums {
		if name == e.Name {
			return &e, true
		}
	}
	for _, e := range enummProtoInfo.enums {
		if name == e.Name {
			return &e, true
		}
	}

	return nil, false
}
func isMessageInCommProtoFile(name string) bool {
	for _, m := range commonProtoInfo.msgs {
		if name == m.Name {
			return true
		}
	}
	for _, m := range enummProtoInfo.msgs {
		if name == m.Name {
			return true
		}
	}

	return false
}
func isMessageInLocalProtoFile(name string, info ProtoInfo) bool {
	for _, m := range info.msgs {
		if name == m.Name {
			return true
		}
	}

	return false
}
func isNestedMessage(name string, msg Message) bool {
	for _, m := range msg.Messages {
		if name == m.Name {
			return true
		}
	}

	return false
}

func isBaseMessageType(msg Message) bool {
	//check base type (such account, role,etc.)
	//check base type
	for _, bs := range FixTables {
		if msg.Name == bs {
			return true
		}
	}
	return false
}
func isBlobMessageType(msg Message) (string, bool) {
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
func isInOrOutMessageType(msg Message) (string, bool) {
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
func isPubMessageType(msg Message) (string, bool) {
	//check pub message, message feature: PUB prefix, both EntityType and UUID exist
	//message will be generated to tcaplusdb table
	flag := checkMessageFlag(msg)
	if strings.HasPrefix(msg.Name, "PUB_") && flag == 2 {
		return "PUB", true
	}
	return "", false
}

func checkMessageFlag(msg Message) int {
	flag := 0
	for _, field := range msg.Fields {
		if field.Name == "EntityType" {
			flag = 1
			continue
		}
		if field.Name == "UUID" {
			flag = flag + 1
			continue
		}
	}
	return flag
}
