package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const protoSuffix = ".proto"

func CreateDir(path string) error {
	fi, err := os.Stat(path)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0744); err != nil {
			return fmt.Errorf("%s create error: %v", path, err)
		}
	}

	if !fi.IsDir() {
		return fmt.Errorf("%s is not directory", path)
	}
	return nil
}
func CreateFile(path string, filename string) (string, error) {
	fi, err := os.Stat(path)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0744); err != nil {
			return "", fmt.Errorf("%s create error: %v", path, err)
		}
	}

	if !fi.IsDir() {
		return "", fmt.Errorf("%s is not directory", path)
	}

	file := filepath.Join(path, filename)
	return file, nil
}

func WriteFile(file string, data []byte) error {

	if err := ioutil.WriteFile(file, data, 0744); err != nil {
		return err
	}

	fmt.Printf("Generated proto: %s\n", file)
	return nil
}
func CheckFile(path string) error {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		return fmt.Errorf("error: %s not exist", path)
	} else if err != nil {
		return fmt.Errorf("error: create fail, %s", err)
	}
	return nil
}

//AbCd => a_b_c_d
//Out_ChatSkins => out_chat_skins
func SnakeCase(str string) string {
	in := []rune(str)
	isLower := func(idx int) bool {
		return idx >= 0 && idx < len(in) && unicode.IsLower(in[idx])
	}
	out := make([]rune, 0, len(in)+len(in)/2)
	for i, r := range in {
		if unicode.IsUpper(r) {
			r = unicode.ToLower(r)
			if i > 0 && in[i-1] != '_' && (isLower(i-1) || isLower(i+1)) {
				out = append(out, '_')
			}
		}
		out = append(out, r)
	}
	return string(out)
}

//traverse proto files of specified directory
func GetProtoFiles(root string, ignores string) ([]string, error) {
	protoFiles := []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// if not a .proto file, do not attempt to parse.
		if !strings.HasSuffix(info.Name(), ".proto") {
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
