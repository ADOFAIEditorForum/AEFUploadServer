package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var suffixMap = map[string]string{
	".png": "image/png",
	".jpg": "image/jpeg", ".jpeg": "image/jpeg",
	".ogg": "audio/ogg", ".mp3": "audio/mpeg",
	".mp4": "application/mp4", ".wav": "audio/wav",
}

func detectMIMEType(fileName string) string {
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
		MaxIdleConns:        100,
	}

	client := &http.Client{
		Transport: transport,
	}

	uploadFiles(url, directory, prefix, client)
	client.CloseIdleConnections()
}

func uploadFiles(url string, directory string, prefix string, client *http.Client) {
	println(directory)
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	i := 0
	fileLength := len(files)
	for i < fileLength {
		file := files[i]

		fileName := file.Name()
		if file.IsDir() {
			uploadFiles(url, filepath.Join(directory, fileName), prefix+fileName+"/", client)
			continue
		}

		if strings.HasSuffix(fileName, ".adofai") {
			continue
		}

		mimeType := detectMIMEType(fileName)
		if mimeType == "" {
			println("Invalid File Type: " + fileName)
			continue
		}

		reqBody, err := os.ReadFile(filepath.Join(directory, fileName))
		if err != nil {
			log.Fatal(err)
		}

		request, err := http.NewRequest("POST", url+"/"+prefix+fileName, bytes.NewBuffer(reqBody))
		if request == nil {
			continue
		}

		request.Close = true
		request.Header.Set("Content-Type", mimeType)
		response, err := client.Do(request)
		if err != nil {
			log.Fatal(err)
			return
		}

		if err != nil {
			log.Fatal(err)
			return
		}

		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
			return
		}

		err = response.Body.Close()
		if err != nil {
			log.Fatal(err)
			return
		}

		println(string(responseBody))
		if string(responseBody) != "Success" {
			println(prefix + fileName)
		} else {
			time.Sleep(10 * time.Millisecond)
		}

		i++
	}
}
