package zip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Encode 压缩文件
// @filePath,文件路径
// @zipPath,储存路劲
func Encode(filePath, zipPath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return compareZip(file, zipWriter, "", !strings.HasSuffix(filePath, "/"))

}

// 压缩文件
func compareZip(file *os.File, zipWriter *zip.Writer, prefix string, join bool) error {
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	if join {
		header.Name = filepath.Join(prefix, header.Name)
		prefix = filepath.Join(prefix, fileInfo.Name())
		//header.Name = strings.ReplaceAll(header.Name, "\\", "/")
		if fileInfo.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate //压缩的关键
		}
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			_, err = io.Copy(writer, file)
			return err
		}
	}

	fileInfoChildList, err := file.Readdir(-1)
	if err != nil {
		return err
	}

	for _, fileInfoChild := range fileInfoChildList {
		fileChild, err := os.Open(filepath.Join(file.Name(), fileInfoChild.Name()))
		if err != nil {
			return err
		}
		if err := compareZip(fileChild, zipWriter, prefix, true); err != nil {
			return err
		}
	}
	return nil

}
