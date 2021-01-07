package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tencentyun/proto-parse-tcaplus/tools"
)

func main() {
	basePath, _ := os.Getwd()
	srcPath := filepath.Join(basePath, "testdata", "yoozoo")
	dstPath := filepath.Join(basePath, "out", "yoozoo")
	if err := tools.CreateDir(dstPath); err != nil {
		fmt.Println(err)
		return
	}
	ProtoParseAndWrite(srcPath, dstPath, IgnoreProtoFiles)
}
