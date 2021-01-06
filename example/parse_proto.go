package example

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/emicklei/proto"
)

const protoSuffix = ".proto"

//global variables
var (
	buf bytes.Buffer
)

func getProtoFiles(root string, ignores string) ([]string, error) {
	protoFiles := []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// if not a .proto file, do not attempt to parse.
		if !strings.HasSuffix(info.Name(), protoSuffix) {
			return nil
		}

		// skip to next if is a directory
		if info.IsDir() {
			return nil
		}

		// skip if path is within an ignored path
		if ignores != "" {
			for _, ignore := range strings.Split(ignores, ",") {
				rel, err := filepath.Rel(filepath.Join(root, ignore), path)
				if err != nil {
					return nil
				}

				if !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
					return nil
				}
			}
		}
		protoFiles = append(protoFiles, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return protoFiles, nil
}

func traverseProtoFiles(basePath string, ignores string, dstPath string) error {
	protoFiles, err := getProtoFiles(basePath, ignores)
	if err != nil {
		return fmt.Errorf("get proto files error : %v", err)

	}
	for _, file := range protoFiles {
		filename := path.Base(file)
		dstFile := filepath.Join(dstPath, filename)
		parse(file, dstFile)
	}
	return nil
}

func parse(protoSrcFile string, protoDstPath string) {

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
	err := tools.writeFile(protoDstPath, buf.Bytes())
	if err != nil {
		fmt.Println(err)
		return
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
	buf.WriteString(`syntax="` + s.Value + `";` + "\n")
}
func handleImport(p *proto.Import) {

	writeComments(p.Comment)
	buf.WriteString(`import "`)
	buf.WriteString(p.Filename)
	buf.WriteString(`";`)
	writeComments(p.InlineComment)

}

func handlePackage(p *proto.Package) {
	writeComments(p.Comment)
	buf.WriteString("package ")

	buf.WriteString(p.Name)

	buf.WriteString("; ")
	writeComments(p.InlineComment)

}

func handleOption(o *proto.Option) {
	if _, ok := o.Parent.(*proto.Proto); !ok {
		//if the message is nested
		return
	}
	writeComments(o.Comment)
	buf.WriteString("option " + o.Name + " = ")
	retStr := parseOptions(o)
	buf.WriteString(*retStr + ";")
	writeComments(o.InlineComment)
}

func handleEnum(e *proto.Enum) {
	if _, ok := e.Parent.(*proto.Proto); !ok {
		return
	}
	parseEnum(e)
}
func parseEnum(e *proto.Enum) {
	writeComments(e.Comment)
	buf.WriteString("enum ")
	buf.WriteString(e.Name)
	buf.WriteString(" { \n ")

	for _, v := range e.Elements {
		//handle enum option

		if eo, ok := v.(*proto.Option); ok {
			writeComments(eo.Comment)
			buf.WriteString("option " + eo.Name + " = ")
			retStr := parseOptions(eo)
			buf.WriteString(*retStr + ";")
			writeComments(eo.InlineComment)
		}

		//handle enum field
		if ef, ok := v.(*proto.EnumField); ok {

			writeComments(ef.Comment)
			writeEnumField(ef)
			writeComments(ef.InlineComment)
		}
		//handle enum reserved
		if r, ok := v.(*proto.Reserved); ok {

			writeComments(r.Comment)
			writeReserved(r)
			writeComments(r.InlineComment)
		}
	}

	buf.WriteString(" }\n")
}
func handleMessage(m *proto.Message) {
	if _, ok := m.Parent.(*proto.Proto); !ok {
		//if the message is nested
		return
	}
	parseMessage(m)
}
func parseMessage(m *proto.Message) {
	writeComments(m.Comment)
	buf.WriteString("message ")
	buf.WriteString(m.Name)
	buf.WriteString(" {\n")
	for _, v := range m.Elements {
		if mo, ok := v.(*proto.Option); ok {
			writeComments(mo.Comment)
			buf.WriteString("option " + mo.Name + " = ")
			retStr := parseOptions(mo)
			buf.WriteString(*retStr + ";")
			writeComments(mo.InlineComment)
		}
		if mnf, ok := v.(*proto.NormalField); ok {
			writeComments(mnf.Comment)
			writeMessageNormalField(mnf)
			writeComments(mnf.InlineComment)
		}

		if mmp, ok := v.(*proto.MapField); ok {

			writeComments(mmp.Comment)
			writeMessageMapField(mmp)
			writeComments(mmp.InlineComment)
		}

		if moo, ok := v.(*proto.Oneof); ok {
			writeComments(moo.Comment)
			writeMessageOneof(moo)
		}

		if r, ok := v.(*proto.Reserved); ok {
			writeComments(r.Comment)
			writeReserved(r)
			writeComments(r.InlineComment)
		}

		if m, ok := v.(*proto.Message); ok {
			parseMessage(m)
		}
		if e, ok := v.(*proto.Enum); ok {
			parseEnum(e)
		}
	}

	buf.WriteString(" }\n")
}

func writeComments(comment *proto.Comment) {
	if comment == nil {
		buf.WriteString("\n")
		return
	}
	for _, line := range comment.Lines {
		buf.WriteString("// ")
		buf.WriteString(line)
		buf.WriteString("\n")
	}
}
func writeEnumField(ef *proto.EnumField) {
	strVal := strconv.Itoa(ef.Integer)
	buf.WriteString(ef.Name + " = " + strVal)
	writeEnumFieldOption(ef)
	buf.WriteString(";")
}
func writeEnumFieldOption(ef *proto.EnumField) {
	fieldOptionStr := ""
	for _, ee := range ef.Elements {
		if o, ok := ee.(*proto.Option); ok {
			retStr := parseOptions(o)
			if fieldOptionStr == "" {
				fieldOptionStr = fmt.Sprintf("%s = %s", o.Name, *retStr)
			} else {
				fieldOptionStr = fmt.Sprintf("%s , %s = %s", fieldOptionStr, o.Name, *retStr)
			}
		}
	}

	if len(ef.Elements) > 0 && fieldOptionStr != "" {
		buf.WriteString(" [ ")
		buf.WriteString(fieldOptionStr)
		buf.WriteString(" ] ")
	}

}

func writeReserved(r *proto.Reserved) {
	buf.WriteString("reserved ")
	rngStr := ""
	fnStr := ""
	// collect all reserved field IDs from the ranges
	for _, rng := range r.Ranges {
		// if range is only a single value, append single value to message's reserved slice
		if rng.From == rng.To {
			if rngStr == "" {
				rngStr = fmt.Sprintf("%d", rng.From)
			} else {
				rngStr = fmt.Sprintf("%s, %d", rngStr, rng.From)
			}
		} else {
			if rngStr == "" {
				rngStr = fmt.Sprintf("%d to %d", rng.From, rng.To)
			} else {
				rngStr = fmt.Sprintf("%s, %d to %d", rngStr, rng.From, rng.To)
			}
		}
	}
	for _, fn := range r.FieldNames {
		if fnStr == "" {
			fnStr = fmt.Sprintf("\"%s\"", fn)
		} else {
			fnStr = fmt.Sprintf("%s, \"%s\"", fnStr, fn)
		}
	}
	if rngStr != "" {
		buf.WriteString(rngStr)
	}
	if fnStr != "" {
		buf.WriteString(fnStr)
	}
	buf.WriteString(";")
}

func parseOptions(o *proto.Option) *string {
	var pOptStr *string
	return recurseLiteral(o.Name, &o.Constant, pOptStr)

}

func recurseLiteral(name string, lit *proto.Literal, pOptStr *string) *string {
	if lit.Map != nil {
		mapStr := ""
		for k, v := range lit.Map {
			retStr := recurseLiteral(k, v, pOptStr)
			if mapStr == "" {
				mapStr = fmt.Sprintf("{ %s : %s", k, *retStr)
			} else {
				mapStr = fmt.Sprintf("%s, %s : %s", mapStr, k, *retStr)
			}
		}
		mapStr = fmt.Sprintf("%s }", mapStr)
		if pOptStr == nil {
			pOptStr = &mapStr
		} else {
			*pOptStr = fmt.Sprintf("%s %s", *pOptStr, mapStr)
		}
		return pOptStr
	}
	if lit.OrderedMap != nil {
		omapStr := ""
		for _, l := range lit.OrderedMap {
			retStr := recurseLiteral(l.Name, l.Literal, pOptStr)
			if omapStr == "" {
				omapStr = fmt.Sprintf("{ %s : %s", l.Name, *retStr)
			} else {
				omapStr = fmt.Sprintf("%s, %s : %s", omapStr, l.Name, *retStr)
			}
		}
		omapStr = fmt.Sprintf("%s }", omapStr)
		if pOptStr == nil {
			pOptStr = &omapStr
		} else {
			*pOptStr = fmt.Sprintf("%s %s", *pOptStr, omapStr)
		}

		return pOptStr
	}

	if lit.Array != nil {
		arrStr := ""
		for _, l := range lit.Array {
			if arrStr == "" {
				arrStr = fmt.Sprintf("[ %v", fmtLiteral(l))
			} else {
				arrStr = fmt.Sprintf("%s, %v", arrStr, fmtLiteral(l))
			}
		}
		arrStr = fmt.Sprintf("%s ]", arrStr)
		if pOptStr == nil {
			pOptStr = &arrStr
		} else {
			*pOptStr = fmt.Sprintf("%s %s", *pOptStr, arrStr)
		}

		return pOptStr
	}
	if pOptStr == nil {
		retStr := fmtLiteral(lit)
		pOptStr = &retStr
	} else {
		*pOptStr = fmt.Sprintf("%s %s", *pOptStr, fmtLiteral(lit))
	}
	return pOptStr
}

func fmtLiteral(liter *proto.Literal) string {
	fmtStr := ""
	if liter.IsString {
		if liter.QuoteRune == 34 {
			fmtStr = fmt.Sprintf("\"%s\"", liter.Source)
		} else if liter.QuoteRune == 39 {
			fmtStr = fmt.Sprintf("'%s'", liter.Source)
		}
	} else {
		fmtStr = fmt.Sprintf("%v", liter.Source)
	}
	return fmtStr
}
func writeMessageOneof(moo *proto.Oneof) {
	buf.WriteString("oneof ")
	buf.WriteString(moo.Name)
	buf.WriteString(" { \n")
	for _, el := range moo.Elements {
		if f, ok := el.(*proto.OneOfField); ok {
			writeComments(f.Comment)
			writeMessageOneofField(f)

		}
	}
	buf.WriteString(" }\n")
}
func writeMessageOneofField(mof *proto.OneOfField) {
	buf.WriteString(mof.Type + " " + mof.Name + " = " + strconv.Itoa(mof.Sequence))
	writeMessageFieldOption(mof.Options)
	buf.WriteString(";\n")
}

func writeMessageNormalField(mnf *proto.NormalField) {
	if mnf.Repeated {
		buf.WriteString("repeated ")
	}
	buf.WriteString(mnf.Type + " " + mnf.Name + " = " + strconv.Itoa(mnf.Sequence))
	writeMessageFieldOption(mnf.Options)
	buf.WriteString(";")
}

func writeMessageMapField(mmp *proto.MapField) {
	f := mmp.Field
	buf.WriteString("map<" + mmp.KeyType + "," + f.Type + "> " + f.Name + " = " + strconv.Itoa(f.Sequence))
	writeMessageFieldOption(mmp.Options)
	buf.WriteString(";")
}
func writeMessageFieldOption(options []*proto.Option) {
	fieldOptionStr := ""
	for _, opt := range options {

		retStr := parseOptions(opt)
		if fieldOptionStr == "" {
			fieldOptionStr = fmt.Sprintf("%s = %s", opt.Name, *retStr)
		} else {
			fieldOptionStr = fmt.Sprintf("%s , %s = %s", fieldOptionStr, opt.Name, *retStr)
		}

	}

	if len(options) > 0 && fieldOptionStr != "" {
		buf.WriteString(" [ ")
		buf.WriteString(fieldOptionStr)
		buf.WriteString(" ] ")
	}

}
