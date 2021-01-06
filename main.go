package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	basePath, _ := os.Getwd()
	dstPath := filepath.Join(basePath, "out", "yoozoo")
	if err := tools.createDir(dstPath); err != nil {
		fmt.Println(err)
		return
	}

}
