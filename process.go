package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
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

var pathMap = map[rune]int{
	'R': 0,
	'p': 15,
	'J': 30,
	'E': 45,
	'T': 60,
	'o': 75,
	'U': 90,
	'q': 105,
	'G': 120,
	'Q': 135,
	'H': 150,
	'W': 165,
	'L': 180,
	'x': 195,
	'N': 210,
	'Z': 225,
	'F': 240,
	'V': 255,
	'D': 270,
	'Y': 285,
	'B': 300,
	'C': 315,
	'M': 330,
	'A': 345,
	'!': 999,
}

var vertexMap = map[rune]struct {
	int
	bool
}{
	'5': {5, false},
	'6': {5, true},
	'7': {7, false},
	'8': {7, true},
}

func getVertex(path rune) (int, bool) {
	result := vertexMap[path]
	return result.int, result.bool
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
	err = os.WriteFile("log.txt", []byte(adofaiLevelStr), 0644)
	if err != nil {
		return
	}

	var adofaiLevelJson map[string]interface{}
	err = json.Unmarshal([]byte(adofaiLevelStr), &adofaiLevelJson)

	if err != nil {
		log.Fatal(err)
		return
	}

	if val, ok := adofaiLevelJson["pathData"]; ok {
		var angleData []float32
		pathData := val.(string)
		for _, path := range pathData {
			if angle, ok := pathMap[path]; ok {
				angleData = append(angleData, float32(angle))
			} else {
				vertex, reverse := getVertex(path)
				vertexCalc := float32(vertex)

				relativeAngle := 180.0 - 180.0*(vertexCalc-2)/vertexCalc
				if reverse {
					relativeAngle = -relativeAngle
				}

				angleData = append(angleData, angleData[len(angleData)-1]+relativeAngle)
			}
		}

		delete(adofaiLevelJson, "pathData")
		adofaiLevelJson["angleData"] = angleData
	}

	jsonData, err := json.Marshal(adofaiLevelJson)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = os.WriteFile(filepath.Join(dest, adofaiFileName), jsonData, 0644)
	if err != nil {
		return
	}

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
