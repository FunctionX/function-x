package utils

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path"
)

func UnCompressZIP(filePath string) (err error) {

	if PathExists(filePath) == false {
		return errors.New("file not exist")
	}

	if IsZip(filePath) == false {
		return errors.New("please upload the zip type file")
	}

	reader, err := zip.OpenReader(filePath)
	defer reader.Close()
	if err != nil {
		return err
	}

	fileDir := path.Dir(filePath)
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			os.Mkdir(path.Join(fileDir, file.Name), file.FileInfo().Mode())
			continue
		}
		// fmt.Println(file.FileInfo().Name(), file.FileInfo().IsDir(), file.FileInfo().Mode(), file.FileInfo().ModTime(), file.FileInfo().Size())
		if err := SaveUnZipFile(file, fileDir); err != nil {
			return err
		}
	}
	return nil
}

func SaveUnZipFile(file *zip.File, fileDir string) (err error) {
	rc, err := file.Open()
	defer rc.Close()
	if err != nil {
		return err
	}
	newFile, err := os.Create(path.Join(fileDir, file.Name))
	defer newFile.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(newFile, rc)
	if err != nil {
		return err
	}
	// fmt.Println(n)
	return nil
}

func IsZip(zipPath string) bool {
	f, err := os.Open(zipPath)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 4)
	if n, err := f.Read(buf); err != nil || n < 4 {
		return false
	}

	return bytes.Equal(buf, []byte("PK\x03\x04"))
}
