package util

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var byteUnits = []string{"B", "KB", "MB", "GB", "TB", "PB"}

func ByteUnitString(n int64) string {
	var unit string
	size := float64(n)
	for i := 1; i < len(byteUnits); i++ {
		if size < 1000 {
			unit = byteUnits[i-1]
			break
		}

		size = size / 1000
	}

	return fmt.Sprintf("%.3g %s", size, unit)
}

func FileExist(path string) bool {
	fi, err := os.Stat(path)
	if err == nil && fi.Size() > 0 {
		return true
	}
	return false
}

func FileWrite(b []byte, filename string) (err error) {
	if len(b) <= 0 {
		return nil
	}
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, b)
	fp.Write(buf.Bytes())
	//log.Printf("save as  %s  (%s)\n", filename, ByteUnitString(int64(len(buf.Bytes()))))
	return nil
}

func FileExt(uri string) string {
	ext := ""
	k := len(uri)
	for i := k - 1; i >= 0; i-- {
		if uri[i] == '?' {
			k = i
			continue
		}
		if uri[i] == '.' {
			ext = uri[i:k]
			break
		}
	}
	return ext
}

func FileName(uri string) string {
	if strings.Contains(uri, "?") {
		pos := strings.Index(uri, "?")
		uri = uri[:pos]
	}
	if strings.Contains(uri, "&") {
		pos := strings.Index(uri, "&")
		uri = uri[:pos]
	}
	name := ""
	for i := len(uri) - 1; i >= 0; i-- {
		if uri[i] == '/' {
			name = uri[i+1:]
			break
		}
	}
	return name
}

func SanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"\\", "_",
		"/", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		"\r", " ",
		"\n", " ",
		"\t", " ",
	)
	name = replacer.Replace(strings.TrimSpace(name))
	name = strings.Join(strings.Fields(name), " ")
	name = strings.Trim(name, ". ")
	if name == "" {
		return "bookget"
	}
	return name
}

func UniqueFilePath(dir string, filename string) string {
	filename = SanitizeFileName(filename)
	fullPath := filepath.Join(dir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fullPath
	}

	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	for i := 1; ; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s(%d)%s", base, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func RenameFileUnique(srcPath string, destDir string, filename string) (string, error) {
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return "", err
	}
	targetPath := UniqueFilePath(destDir, filename)
	if srcPath == targetPath {
		return targetPath, nil
	}
	if err := os.Rename(srcPath, targetPath); err != nil {
		return "", err
	}
	return targetPath, nil
}

// 压缩文件
func Zip(srcFile string, destZip string) error {
	zipFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	archive := zip.NewWriter(zipFile)
	defer archive.Close()
	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = path
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err = io.Copy(writer, file); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// 解压缩
func Unzip(zipFile string, destDir string, sortId string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()
	for _, f := range zipReader.File {
		newName := f.Name
		if sortId != "" {
			newName = fmt.Sprintf("%s.%s", sortId, f.Name)
		}
		fpath := filepath.Join(destDir, newName)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}
			inFile, err := f.Open() //这个是从压缩文件读取出来的
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()) //创建的新文件
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err = io.Copy(outFile, inFile); err != nil {
				return err
			}
		}
	}
	return nil
}
