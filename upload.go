package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var suffixMap = map[string]string{
	".png": "image/png",
	".jpg": "image/jpeg", ".jpeg": "image/jpeg",
	".ogg": "audio/ogg", ".mp3": "audio/mpeg",
	".mp4": "application/mp4", ".wav": "audio/wav",
}

func detectMIMEType(fileName string) string {
	fileName = strings.ToLower(fileName)
	for suffix, mimeType := range suffixMap {
		if strings.HasSuffix(fileName, suffix) {
			return mimeType
		}
	}

	return ""
}

func uploadAll(url string, directory string, prefix string) {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 40,
		MaxConnsPerHost:     40,
		MaxIdleConns:        80,
	}

	client := &http.Client{
		Transport: transport,
	}

	log.Println(url)

	filesDetected := uploadFiles(url, directory, prefix, client, 0, 0)
	log.Println(fmt.Sprintf("UPLOAD COMPLETE | %d files", filesDetected))
	client.CloseIdleConnections()

	err := os.RemoveAll(directory)
	if err != nil {
		log.Fatal(err)
	}
}

func uploadFiles(link string, directory string, prefix string, client *http.Client, filesDetected int32, depth int32) int32 {
	log.Println(directory)
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	i := 0
	fileLength := len(files)
	log.Println(fmt.Sprintf("%d files", fileLength))

	for i < fileLength {
		file := files[i]

		fileName := file.Name()
		if file.IsDir() {
			filesDetected += uploadFiles(link, filepath.Join(directory, fileName), prefix+fileName+"/", client, 0, depth+1)
			i++

			continue
		}

		if strings.HasSuffix(fileName, ".adofai") {
			i++

			continue
		}

		mimeType := detectMIMEType(fileName)
		if mimeType == "" {
			println("Invalid File Type: " + fileName)
			i++

			continue
		}

		reqBody, err := os.ReadFile(filepath.Join(directory, fileName))
		if err != nil {
			log.Fatal(err)
		}

		request, err := http.NewRequest("POST", link, bytes.NewBuffer(reqBody))

		if request == nil || err != nil {
			log.Fatal(err)
		}

		// request.Close = true
		request.Header.Set("Content-Type", mimeType)
		request.Header.Set("File-Path", url.QueryEscape(prefix+fileName))
		response, err := client.Do(request)
		if err != nil {
			log.Fatal(err)
		}

		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = response.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		println(prefix + fileName)
		println(string(responseBody))
		if string(responseBody) == "Success" {
			filesDetected++
			// time.Sleep(10 * time.Millisecond)
		}

		i++
	}

	if depth <= 0 {
		emptyData := []byte{}
		request, err := http.NewRequest("DELETE", link, bytes.NewBuffer(emptyData))
		if err != nil {
			log.Fatal(err)
		}

		response, err := client.Do(request)
		if err != nil {
			log.Fatal(err)
		}

		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = response.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		println(string(responseBody))

		var uploadInfo map[string]interface{}
		err = json.Unmarshal(responseBody, &uploadInfo)
		if err != nil {
			log.Fatal(err)
		}
	}

	return filesDetected
}
