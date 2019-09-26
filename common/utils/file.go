package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
)

// check path exist
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

//pwd
func GetCurPath() (s string) {
	if pwd, err := os.Getwd(); err == nil {
		return pwd
	}
	return
}

func GetExePath() (s string) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		return dir
	}
	return
}

func CheckMkdir(path string) error {
	if PathExists(path) {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

func CreateFile(name string) error {
	fo, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		_ = fo.Close()
	}()
	return nil
}

func DirFolders(dir string) (folders []string) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, v := range infos {
		if v.IsDir() {
			folders = append(folders, v.Name())
		}
	}
	return
}

func LogFile(filePath string) io.Writer {
	dir := path.Dir(filePath)
	if !PathExists(dir) {
		_ = os.MkdirAll(dir, os.ModePerm)
	}
	// os.Remove(filePath)
	if !PathExists(filePath) {
		_ = CreateFile(filePath)
	}
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("============== create file log fatal", err.Error())
		return os.Stdout
	}
	return f
}

func GetDirName(dir string, isFirstDir bool) (names map[string][]string) {
	names = map[string][]string{}
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, v := range infos {
		if v.IsDir() == false {
			continue
		}

		if isFirstDir {
			names[v.Name()] = []string{}
		} else {
			names[path.Base(dir)] = append(names[path.Base(dir)], v.Name())
		}
	}
	if isFirstDir == false {
		return names
	}
	for k := range names {
		names[k] = GetDirName(path.Join(dir, k), false)[k]
	}

	return names
}

// load json from file, and parse to val
func LoadJSON(file string, val interface{}) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, val); err == nil {
		return nil
	}
	if syntaxerr, ok := err.(*json.SyntaxError); ok {
		line := findLine(content, syntaxerr.Offset)
		return fmt.Errorf("JSON syntax error at %v:%v: %v", file, line, err)
	}
	return fmt.Errorf("JSON unmarshal error in %v: %v", file, err)
}

// findLine returns the line number for the given offset into data.
func findLine(data []byte, offset int64) (line int) {
	line = 1
	for i, r := range string(data) {
		if int64(i) >= offset {
			return
		}
		if r == '\n' {
			line++
		}
	}
	return
}

func SaveJSON(file string, val interface{}) error {
	bytes, err := json.MarshalIndent(val, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, bytes, os.ModePerm)
}

func MakeReaderFromPath(path string, fieldName string, params map[string]string) (string, io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()
	return MakeReaderFromFile(file, fieldName, filepath.Base(path), params)
}

func MakeReaderFromFile(reader io.Reader, fieldName string, fileName string, params map[string]string) (string, io.Reader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return "", nil, err
	}
	_, err = io.Copy(part, reader)
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return "", nil, err
	}
	return writer.FormDataContentType(), body, nil
}
