package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var defaultFileList = []string{"level.adofai", "main.adofai"}

func detectADOFAIFile(destination string) string {
	adofaiFileName := ""
	files, err2 := os.ReadDir(destination)
	if err2 != nil {
		log.Fatal(err2)
		return ""
	}

	for _, file := range files {
		name := file.Name()

		if strings.HasSuffix(name, ".adofai") {
			if slices.Contains(defaultFileList, name) {
				adofaiFileName = name
				break
			}
			if adofaiFileName == "" && name != "backup.adofai" {
				adofaiFileName = name
			}
		}
	}

	return adofaiFileName
}

func process(filename string, id int64) {
	dest := fmt.Sprintf("level%d", id)

	err := unzipSource(filename, dest)
	if err != nil {
		return
	}

	err = os.Remove(filename)
	if err != nil {
		return
	}

	adofaiFileName := detectADOFAIFile(dest)

	println(adofaiFileName)

	data, err := os.ReadFile(filepath.Join(dest, adofaiFileName))
	trimmedBytes := bytes.Trim(data, "\xef\xbb\xbf")
	adofaiLevelStr := string(trimmedBytes)

	adofaiLevelStr = convertToValidJSON(adofaiLevelStr)
	jsonData := []byte(adofaiLevelStr)

	apiURL := "http://localhost:3677/level"
	request, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		println(err)
		return
	}

	responseBody, err := io.ReadAll(response.Body)
	println(responseBody)

	return
}

func unzipSource(filename string, destination string) error {
	reader, err := zip.OpenReader(filename)

	if err != nil {
		return err
	}
	defer func(reader *zip.ReadCloser) {
		err := reader.Close()
		if err != nil {
			return
		}
	}(reader)

	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		err := unzipFile(f, destination)
		if err != nil {
			return err
		}
	}

	return nil
}

func unzipFile(file *zip.File, destination string) error {
	filePath := filepath.Join(destination, file.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}

		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer func(destinationFile *os.File) {
		err := destinationFile.Close()
		if err != nil {
			return
		}
	}(destinationFile)

	zippedFile, err := file.Open()
	if err != nil {
		return err
	}
	defer func(zippedFile io.ReadCloser) {
		err := zippedFile.Close()
		if err != nil {
			return
		}
	}(zippedFile)

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}
